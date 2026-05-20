package retrieve

import (
	"testing"
)

func TestSplitChunks(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single paragraph",
			input: "hello world",
			want:  []string{"hello world"},
		},
		{
			name:  "two paragraphs",
			input: "first paragraph\n\nsecond paragraph",
			want:  []string{"first paragraph", "second paragraph"},
		},
		{
			name:  "heading boundary",
			input: "intro text\n# Heading\nheading content",
			want:  []string{"intro text", "# Heading\nheading content"},
		},
		{
			name:  "blank-only input",
			input: "\n\n\n",
			want:  nil,
		},
		{
			name:  "multiple blank lines collapse",
			input: "first\n\n\n\nsecond",
			want:  []string{"first", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitChunks(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitChunks() returned %d chunks, want %d\ngot:  %#v\nwant: %#v", len(got), len(tt.want), got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("chunk[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
