package str

import "testing"

func TestCensorText(t *testing.T) {
	profanity = `
# comments

badword|worseword
`

	tests := []struct {
		input    string
		partial  string
		expected string
	}{
		{"this is a badword", "full", "this is a *******"},
		{"a worseword is here", "full", "a ********* is here"},
		{"badword worseword", "partial", "b*****d w*******d"},
		{"badword worseword", "invalid-type", "badword worseword"},
		{"no bad content", "full", "no bad content"},
	}

	for _, test := range tests {
		output := CensorText(test.input, test.partial)
		if output != test.expected {
			t.Errorf("CensorText(%q, %v) = %q; want %q", test.input, test.partial, output, test.expected)
		}
	}
}
