// Package crashlog keeps a program's crashes in a small local file, so a
// panic that kills trackd at night is still there in the morning (and in
// tock bugreport). It writes nothing until a crash and sends nothing.
//
// Install points the Go runtime's fatal output (debug.SetCrashOutput) at
// <state>/<program>-crash.log: every unrecovered panic, in any goroutine,
// and every fatal runtime error lands there as well as on stderr. Recover,
// deferred at the top of a goroutine, adds a line saying when, which
// program and where, then lets the panic go on: the program still dies and
// its supervisor (launchd, systemd) still restarts it.
package crashlog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

// MaxBytes is the size above which Install starts a new file. The old one
// is kept as <file>.1 (replacing the one before), so the pair stays under
// about twice this plus one crash.
const MaxBytes = 256 << 10

// Suffix ends every crash log's name.
const Suffix = "-crash.log"

// Path is program's crash log in stateDir.
func Path(stateDir, program string) string {
	return filepath.Join(stateDir, program+Suffix)
}

var (
	mu      sync.Mutex
	out     *os.File
	program string
	version string
)

// Install opens program's crash log (0600, in a 0700 dir), starting a new
// one if it is over MaxBytes, and makes the runtime write fatal errors to
// it. It is called once, early in main; a second call replaces the first.
func Install(stateDir, prog, ver string) error {
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return fmt.Errorf("crashlog: %w", err)
	}
	path := Path(stateDir, prog)
	if err := rotate(path, MaxBytes); err != nil {
		return fmt.Errorf("crashlog: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("crashlog: %w", err)
	}
	if err := f.Chmod(0o600); err != nil && !errors.Is(err, errors.ErrUnsupported) {
		f.Close()
		return fmt.Errorf("crashlog: %w", err)
	}
	if err := debug.SetCrashOutput(f, debug.CrashOptions{}); err != nil {
		f.Close()
		return fmt.Errorf("crashlog: %w", err)
	}
	mu.Lock()
	old := out
	out, program, version = f, prog, ver
	mu.Unlock()
	if old != nil {
		old.Close()
	}
	return nil
}

// rotate moves path to path.1 when it is bigger than max. A failure to move
// (Windows, while another process holds it open) is not an error: the file
// only grows when a program crashes, and the next start tries again.
func rotate(path string, max int64) error {
	fi, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if fi.Size() <= max {
		return nil
	}
	_ = os.Rename(path, path+".1")
	return nil
}

// Recover is deferred at the top of a goroutine (defer crashlog.Recover(
// "sync loop")). On a panic it writes one line, the time, program,
// version, pid and where, then panics again with the same value, so the
// runtime writes the panic and its stack under that line and the program
// exits as it would have without it.
func Recover(where string) {
	r := recover()
	if r == nil {
		return
	}
	Note(where)
	panic(r)
}

// Note writes Recover's line without recovering anything, for a caller
// that handles the panic itself.
func Note(where string) {
	mu.Lock()
	defer mu.Unlock()
	if out == nil {
		return
	}
	fmt.Fprintf(out, "\n--- %s %s %s pid %d: panic in %s\n",
		time.Now().UTC().Format(time.RFC3339), program, version, os.Getpid(), where)
}
