package tui

import "testing"

func TestTruncate(t *testing.T) {
	cases := []struct {
		in    string
		width int
		want  string
	}{
		{"i-0123456789", 20, "i-0123456789"},
		{"i-0123456789abcdef", 8, "i-01234…"},
		{"abc", 0, ""},
		{"abc", 1, "a"},
	}

	for _, c := range cases {
		if got := truncate(c.in, c.width); got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.width, got, c.want)
		}
	}
}

func TestGaugeColorsSeverity(t *testing.T) {
	low, _ := gaugeColors(0.2)
	mid, _ := gaugeColors(0.7)
	high, _ := gaugeColors(0.95)

	if low != hexGreen {
		t.Errorf("expected low ratio to start green, got %s", low)
	}
	if mid != hexViolet {
		t.Errorf("expected mid ratio to start violet, got %s", mid)
	}
	if high != hexSky {
		t.Errorf("expected high ratio to start sky, got %s", high)
	}
}
