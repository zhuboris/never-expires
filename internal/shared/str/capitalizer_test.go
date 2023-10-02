package str

import "testing"

func TestCapitalized(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "not capitalized string",
			input: "test",
			want:  "Test",
		},
		{
			name:  "already capitalized string",
			input: "Test",
			want:  "Test",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "one lower character",
			input: "a",
			want:  "A",
		},
		{
			name:  "one capital character",
			input: "B",
			want:  "B",
		},
		{
			name:  "string starting with space",
			input: " test",
			want:  " test",
		},
		{
			name:  "string starting with not letter",
			input: "%test",
			want:  "%test",
		},
		{
			name:  "string starting with unicode symbol",
			input: "😃test",
			want:  "😃test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := Capitalized(tt.input); result != tt.want {
				t.Errorf("result = %q, want %q", result, tt.want)
			}
		})
	}
}
