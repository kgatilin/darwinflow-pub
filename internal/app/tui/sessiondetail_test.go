package tui_test

import (
	"testing"

	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
)

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "small number",
			count: 123,
			want:  "~123 tokens",
		},
		{
			name:  "one thousand",
			count: 1000,
			want:  "~1,000 tokens",
		},
		{
			name:  "five thousand",
			count: 5432,
			want:  "~5,432 tokens",
		},
		{
			name:  "fifty thousand",
			count: 50000,
			want:  "~50,000 tokens",
		},
		{
			name:  "one hundred thousand",
			count: 123456,
			want:  "~123,456 tokens",
		},
		{
			name:  "one million",
			count: 1000000,
			want:  "~1,000,000 tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.FormatTokenCount(tt.count)
			if got != tt.want {
				t.Errorf("FormatTokenCount(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}
