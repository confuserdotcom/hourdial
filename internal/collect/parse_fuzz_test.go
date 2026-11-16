package collect

import (
	"bytes"
	"testing"
)

// Replies from the window manager are input from another program: a
// broken or hostile one may send anything, and tock must not panic or
// allocate without bound.
func FuzzI3Read(f *testing.F) {
	f.Add(i3Encode(4, []byte(`{"nodes":[]}`)))
	f.Add(i3Encode(0x80000003, []byte(`{"change":"focus"}`)))
	f.Add([]byte(i3Magic + "\xff\xff\xff\xff\x00\x00\x00\x00"))
	f.Add([]byte("i3-ip"))
	f.Fuzz(func(t *testing.T, b []byte) {
		_, p, err := i3Read(bytes.NewReader(b))
		if err == nil && len(p) > i3MaxPayload {
			t.Fatalf("payload of %d bytes accepted", len(p))
		}
	})
}

func FuzzI3Tree(f *testing.F) {
	f.Add([]byte(`{"nodes":[{"focused":true,"app_id":"foot","name":"~"}]}`))
	f.Add([]byte(`{"floating_nodes":[{"focused":true,"window_properties":{"class":"Firefox"},"name":"t"}]}`))
	f.Add([]byte(`{"change":"title","container":{"focused":true,"app_id":"a","name":"b"}}`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(t *testing.T, b []byte) {
		focusedWindow(b)
		windowEvent(b)
	})
}

func FuzzHyprland(f *testing.F) {
	f.Add("activewindow>>kitty,~/tock: vim", []byte(`{"class":"kitty","title":"vim"}`))
	f.Add("activewindow>>,", []byte(`{}`))
	f.Add("workspace>>2", []byte(`null`))
	f.Fuzz(func(t *testing.T, line string, reply []byte) {
		if class, _, ok := parseActiveWindowEvent(line); ok && bytes.ContainsRune([]byte(class), ',') {
			t.Fatalf("class %q holds a comma", class)
		}
		parseActiveWindow(reply)
	})
}
