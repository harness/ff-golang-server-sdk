package types

import "testing"

func TestInteger_Equal(t *testing.T) {

	tests := []struct {
		name string
		n    Integer
		args []string
		want bool
	}{
		{"test equal returns true with match", Integer(22), []string{"22"}, true},
		{"test equal returns false when no match", Integer(22), []string{"25"}, false},
		{"test equal returns true with multiple values", Integer(22), []string{"22", "23"}, true},
		{"test equal only matches first value", Integer(22), []string{"23", "22"}, false},
		{"test equal returns false for invalid value", Integer(22), []string{"true"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Equal(tt.args); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
