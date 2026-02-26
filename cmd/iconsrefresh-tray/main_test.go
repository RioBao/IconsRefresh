package main

import (
	"testing"

	"github.com/crazy-max/IconsRefresh/internal/engine"
)

func TestParseTrayPreset(t *testing.T) {
	tests := []struct {
		input string
		want  engine.Preset
		ok    bool
	}{
		{input: "quick", want: engine.PresetTrayQuick, ok: true},
		{input: "standard", want: engine.PresetTrayStandard, ok: true},
		{input: "deep", want: engine.PresetTrayDeep, ok: true},
		{input: "invalid", ok: false},
	}

	for _, tc := range tests {
		got, err := parseTrayPreset(tc.input)
		if tc.ok && err != nil {
			t.Fatalf("parseTrayPreset(%q) unexpected error: %v", tc.input, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("parseTrayPreset(%q) expected error", tc.input)
		}
		if tc.ok && got != tc.want {
			t.Fatalf("parseTrayPreset(%q)=%q want %q", tc.input, got, tc.want)
		}
	}
}
