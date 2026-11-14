package crashlog

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"tock/internal/testfs"
)

// install is Install for a test in this process: it lets go of the file
// at the end (and the runtime's copy of it), so Windows can remove the
// temp dir.
func install(t *testing.T, dir, prog, ver string) {
	t.Helper()
	if err := Install(dir, prog, ver); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		debug.SetCrashOutput(nil, debug.CrashOptions{})
		mu.Lock()
		defer mu.Unlock()
		if out != nil {
			out.Close()
			out = nil
		}
	})
}

// child runs this test binary as a program that crashes the way mode says,
// with its crash log in dir, and returns the log.
func child(t *testing.T, dir, mode string) string {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestCrashChild$")
	cmd.Env = append(os.Environ(), "CRASHLOG_CHILD="+mode, "CRASHLOG_DIR="+dir)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("%s: the child did not crash:\n%s", mode, out)
	}
	if !strings.Contains(string(out), "boom-"+mode) {
		t.Fatalf("%s: stderr lost the panic:\n%s", mode, out)
	}
	b, err := os.ReadFile(Path(dir, "trackd"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestCrashChild is the crashing program child starts; on its own it does
// nothing.
func TestCrashChild(t *testing.T) {
	mode := os.Getenv("CRASHLOG_CHILD")
	if mode == "" {
		t.Skip("run by child")
	}
	if err := Install(os.Getenv("CRASHLOG_DIR"), "trackd", "v9.9.9"); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	switch mode {
	case "wrapped":
		go func() {
			defer Recover("sync loop")
			panic("boom-" + mode)
		}()
	case "bare":
		go func() { panic("boom-" + mode) }()
	case "main":
		defer Recover("main")
		panic("boom-" + mode)
	}
	<-done
}

func TestWrappedGoroutinePanicIsLogged(t *testing.T) {
	testfs.UnixOnly(t, "file modes")
	dir := filepath.Join(t.TempDir(), "state")
	log := child(t, dir, "wrapped")
	for _, want := range []string{"trackd v9.9.9 pid ", "panic in sync loop", "boom-wrapped", "goroutine ", "crashlog_test.go"} {
		if !strings.Contains(log, want) {
			t.Errorf("log has no %q:\n%s", want, log)
		}
	}
	if strings.Index(log, "panic in sync loop") > strings.Index(log, "panic: boom-wrapped") {
		t.Errorf("the note should come before the trace:\n%s", log)
	}
	fi, err := os.Stat(Path(dir, "trackd"))
	if err != nil {
		t.Fatal(err)
	}
	if !testfs.Perm(fi, 0o600) {
		t.Fatalf("mode %v, want 0600", fi.Mode().Perm())
	}
	di, _ := os.Stat(dir)
	if !testfs.Perm(di, 0o700) {
		t.Fatalf("dir mode %v, want 0700", di.Mode().Perm())
	}
}

// A goroutine without Recover, and a panic on the main goroutine, still
// reach the file through the runtime.
func TestUnwrappedPanicsAreLogged(t *testing.T) {
	for _, mode := range []string{"bare", "main"} {
		log := child(t, t.TempDir(), mode)
		if !strings.Contains(log, "panic: boom-"+mode) || !strings.Contains(log, "goroutine ") {
			t.Errorf("%s: log has no panic and stack:\n%s", mode, log)
		}
	}
}

// Two crashes append; nothing is written while nothing crashes.
func TestCrashesAppendAndQuietRunsWriteNothing(t *testing.T) {
	dir := t.TempDir()
	child(t, dir, "wrapped")
	log := child(t, dir, "wrapped")
	if n := strings.Count(log, "panic in sync loop"); n != 2 {
		t.Fatalf("%d notes, want 2:\n%s", n, log)
	}
	quiet := t.TempDir()
	install(t, quiet, "tock", "v1")
	Recover("nothing") // no panic: a no-op
	if fi, err := os.Stat(Path(quiet, "tock")); err != nil || fi.Size() != 0 {
		t.Fatalf("quiet run wrote %v bytes (err %v)", fi.Size(), err)
	}
}

// A log over MaxBytes is moved to .1 at the next start, so the pair
// stays bounded however often the program crashes.
func TestInstallRotates(t *testing.T) {
	dir := t.TempDir()
	path := Path(dir, "tockbar")
	big := strings.Repeat("x", MaxBytes+1)
	if err := os.WriteFile(path, []byte(big), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".1", []byte("older"), 0o600); err != nil {
		t.Fatal(err)
	}
	install(t, dir, "tockbar", "v1")
	if fi, _ := os.Stat(path); fi.Size() != 0 {
		t.Fatalf("new log holds %d bytes, want 0", fi.Size())
	}
	if b, _ := os.ReadFile(path + ".1"); len(b) != len(big) {
		t.Fatalf(".1 holds %d bytes, want the old log", len(b))
	}
	// Under the cap, nothing moves.
	os.WriteFile(path, []byte("small"), 0o600)
	install(t, dir, "tockbar", "v1")
	if b, _ := os.ReadFile(path); string(b) != "small" {
		t.Fatalf("log = %q, want it kept", b)
	}
}
