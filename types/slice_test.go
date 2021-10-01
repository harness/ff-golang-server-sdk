package types

import "testing"

func TestSlice_In(t *testing.T) {

	tests := []struct {
		name   string
		fields interface{}
		args   []string
		want   bool
	}{
		{"test string comparison returns true for match", []string{"alpha", "bravo", "charlie"}, []string{"one", "charlie", "three"}, true},
		{"test string comparison returns false for no match", []string{"alpha", "bravo", "charlie"}, []string{"one", "two", "three"}, false},
		{"test float64 comparison returns true for match", []float64{10, 40, 60}, []string{"20", "50", "40"}, true},
		{"test float64 comparison returns false for no match", []float64{10, 40, 60}, []string{"20", "50", "70"}, false},
		{"test bool comparison returns true for match", []bool{true}, []string{"true"}, true},
		{"test bool comparison returns false for no match", []bool{true}, []string{"false"}, false},
		{"test string interface{} comparison returns true for match", []interface{}{"alpha", "bravo", "charlie"}, []string{"one", "charlie", "three"}, true},
		{"test string interface{} comparison returns false for no match", []interface{}{"alpha", "bravo", "charlie"}, []string{"one", "two", "three"}, false},
		{"test string comparison ignores case", []string{"alpha", "bravo", "charlie"}, []string{"one", "BRAVO", "three"}, true},
		{"test ptr to interface is handled", &[]interface{}{"alpha", "bravo", "charlie"}, []string{"one", "BRAVO", "three"}, true},
		{"test nil returns false", nil, []string{"one", "BRAVO", "three"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Slice{
				data: tt.fields,
			}
			if got := s.In(tt.args); got != tt.want {
				t.Errorf("In() = %v, want %v", got, tt.want)
			}
		})
	}
}
