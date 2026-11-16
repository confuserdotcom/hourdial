package awimport

import (
	"testing"
	"time"
)

// An export is a file from another program. Whatever it holds, Parse
// returns an error or views that run forward in time; it never panics.
func FuzzParse(f *testing.F) {
	f.Add([]byte(export))
	f.Add([]byte(`{"buckets":{"w":{"type":"currentwindow","events":[{"timestamp":"2026-01-01T23:59:00Z","duration":1e300,"data":{"app":"a","title":"t"}}]}}}`))
	f.Add([]byte(`{"buckets":{"w":{"type":"currentwindow","events":[{"timestamp":"0001-01-01T00:00:00Z","duration":-5,"data":{}}]}}}`))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		days, err := Parse(b, time.UTC)
		if err != nil {
			return
		}
		for _, d := range days {
			for _, v := range d.Views {
				if v.EndMS < v.StartMS {
					t.Fatalf("view runs backwards: %+v", v)
				}
			}
		}
	})
}
