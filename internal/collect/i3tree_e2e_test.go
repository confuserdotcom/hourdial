//go:build e2e

package collect

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

// TestRealTreeSize reads the get_tree reply of the running Sway or i3 (the
// e2e scripts open a few dozen windows first) and logs its size per window
// and what decoding it allocates: the reply cap rests on these numbers.
func TestRealTreeSize(t *testing.T) {
	sock := os.Getenv("SWAYSOCK")
	if sock == "" {
		sock = os.Getenv("I3SOCK")
	}
	if sock == "" {
		t.Skip("no Sway or i3 session")
	}
	tree, err := i3Request(sock, i3GetTree, nil)
	if err != nil {
		t.Fatal(err)
	}
	var count func(n *i3Node) int
	count = func(n *i3Node) int {
		c := 0
		if n.class() != "" {
			c++
		}
		for i := range n.Nodes {
			c += count(&n.Nodes[i])
		}
		for i := range n.FloatingNodes {
			c += count(&n.FloatingNodes[i])
		}
		return c
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	if _, _, err := focusedWindow(tree); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	var root i3Node
	if err := json.Unmarshal(tree, &root); err != nil {
		t.Fatal(err)
	}
	windows := count(&root)
	if windows == 0 {
		t.Fatal("no windows in the tree")
	}
	t.Logf("tree-size: windows=%d bytes=%d per_window=%d decode_alloc=%d",
		windows, len(tree), len(tree)/windows, after.TotalAlloc-before.TotalAlloc)
	// With enough windows the fixed part (outputs, workspaces) is small, and
	// the cap must still hold a desktop of 3,000 of them.
	if windows >= 20 && len(tree)/windows*3000 > i3MaxPayload {
		t.Errorf("%d bytes per window: the reply cap of %d bytes holds fewer than 3000 windows", len(tree)/windows, i3MaxPayload)
	}
}
