package task

import "testing"

func TestPriorityString(t *testing.T) {
	cases := []struct {
		name string
		p    Priority
		want string
	}{
		{"low", PriorityLow, "low"},
		{"medium", PriorityMedium, "medium"},
		{"high", PriorityHigh, "high"},
		{"negative", Priority(-1), "unknown"},
		{"out of range", Priority(99), "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.String(); got != tc.want {
				t.Errorf("Priority(%d).String() = %q, want %q", tc.p, got, tc.want)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    Priority
		wantErr bool
	}{
		{"name low", "low", PriorityLow, false},
		{"name uppercase", "HIGH", PriorityHigh, false},
		{"numeric", "2", PriorityHigh, false},
		{"invalid name", "urgent", 0, true},
		{"numeric out of range", "9", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParsePriority(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParsePriority(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePriority(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParsePriority(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
