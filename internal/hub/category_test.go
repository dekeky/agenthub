package hub

import "testing"

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", CategoryPicoClaw, false},
		{"picoclaw", CategoryPicoClaw, false},
		{"OpenClaw", CategoryOpenClaw, false},
		{"openclaw", CategoryOpenClaw, false},
		{"../evil", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizeCategory(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("NormalizeCategory(%q) expected error", tt.in)
			}
			if !IsInvalidCategory(err) {
				t.Fatalf("NormalizeCategory(%q) err = %v, want ErrInvalidCategory", tt.in, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeCategory(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("NormalizeCategory(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
