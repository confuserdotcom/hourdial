package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDisplay(t *testing.T) {
	c, err := Load(write(t, ""))
	if err != nil {
		t.Fatal(err)
	}
	if c.Display.FirstDay() != time.Monday || c.Display.WeekStartName() != "monday" {
		t.Fatalf("default first day = %v", c.Display.FirstDay())
	}
	if (Display{}).FirstDay() != time.Monday {
		t.Fatal("zero Display must start weeks on Monday")
	}
	for in, want := range map[string]time.Weekday{"monday": time.Monday, "Sunday": time.Sunday, " saturday ": time.Saturday} {
		c, err := Load(write(t, "[display]\nweek_start = \""+in+"\"\n"))
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if c.Display.FirstDay() != want {
			t.Fatalf("%q: first day = %v, want %v", in, c.Display.FirstDay(), want)
		}
	}
	for _, bad := range []string{"tuesday", "", "mon"} {
		if _, err := Load(write(t, "[display]\nweek_start = \""+bad+"\"\n")); err == nil || !strings.Contains(err.Error(), "week_start") {
			t.Fatalf("%q: err = %v", bad, err)
		}
	}
	if _, err := Load(write(t, "[display]\nweekstart = \"sunday\"\n")); err == nil || !strings.Contains(err.Error(), "weekstart") {
		t.Fatalf("unknown key: err = %v", err)
	}
}
