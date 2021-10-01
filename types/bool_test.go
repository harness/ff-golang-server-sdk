package types

import "testing"

func TestBoolean_Equal(t *testing.T) {
	tests := []struct {
		name string
		b    Boolean
		args []string
		want bool
	}{
		{"test equal returns true with match", Boolean(true), []string{"true"}, true},
		{"test equal returns false when no match", Boolean(true), []string{"false"}, false},
		{"test equal returns true with multiple values", Boolean(false), []string{"false", "true"}, true},
		{"test equal only matches first value", Boolean(false), []string{"true", "false"}, false},
		{"test equal returns false for invalid value", Boolean(true), []string{"on"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.Equal(tt.args); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
