package finder

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestSanitizeControlStripsEscapes verifies that terminal escape and control
// sequences embedded in untrusted input are neutralized. This is the core
// defense against escape-sequence injection from malicious filenames and file
// contents (an attacker who controls a browsed directory could otherwise set
// the window title, move the cursor, or hijack the clipboard via OSC 52).
func TestSanitizeControlStripsEscapes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"sgr color", "file\x1b[31mred.txt", "file?[31mred.txt"},
		{"osc set-title", "f\x1b]0;pwned\x07.txt", "f?]0;pwned?.txt"},
		{"bare bel", "ding\x07", "ding?"},
		{"carriage return", "a\rb", "a?b"},
		{"newline", "a\nb", "a?b"},
		{"tab", "a\tb", "a?b"},
		{"del", "a\x7fb", "a?b"},
		{"c1 nel", "a\u0085b", "a?b"},
		{"clean ascii", "normal_file.go", "normal_file.go"},
		{"clean unicode", "café-☃.txt", "café-☃.txt"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeControl(tc.in); got != tc.want {
				t.Errorf("sanitizeControl(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	// The ESC byte itself must never survive — it is the prefix for every
	// dangerous CSI/OSC sequence.
	for _, tc := range cases {
		if strings.ContainsRune(sanitizeControl(tc.in), '\x1b') {
			t.Errorf("sanitizeControl(%q) leaked an ESC byte", tc.in)
		}
	}
}

// TestSanitizeControlKeepRunes verifies that callers can preserve specific
// runes (the preview path keeps '\n' and '\t' because clipLines handles them).
func TestSanitizeControlKeepRunes(t *testing.T) {
	in := "line1\n\tindented\x1b[31m\nline2"
	got := sanitizeControl(in, '\n', '\t')
	want := "line1\n\tindented?[31m\nline2"
	if got != want {
		t.Errorf("sanitizeControl(%q, keep \\n \\t) = %q, want %q", in, got, want)
	}
	if strings.ContainsRune(got, '\x1b') {
		t.Error("ESC byte survived sanitization even with keep set")
	}
}

// TestSanitizeControlPreservesLegitNames guards the fast path: a name with no
// control characters must be returned byte-for-byte unchanged, so the fuzzy
// highlight offsets computed against the original name stay aligned.
func TestSanitizeControlPreservesLegitNames(t *testing.T) {
	for _, name := range []string{"main.go", "café.txt", "a b c", "RÉSUMÉ.PDF", "日本語.md"} {
		if got := sanitizeControl(name); got != name {
			t.Errorf("sanitizeControl(%q) = %q, want unchanged", name, got)
		}
	}
}

// TestTruncateTailRuneSafe verifies the breadcrumb's left-truncation keeps the
// trailing portion within the width budget and never splits a multibyte rune.
func TestTruncateTailRuneSafe(t *testing.T) {
	cases := []struct {
		in    string
		width int
		want  string
	}{
		{"/home/user/project", 7, "project"},
		{"/a/café", 4, "café"},
		{"short", 100, "short"},
		{"anything", 0, ""},
		{"anything", -3, ""},
	}
	for _, tc := range cases {
		got := truncateTail(tc.in, tc.width)
		if got != tc.want {
			t.Errorf("truncateTail(%q, %d) = %q, want %q", tc.in, tc.width, got, tc.want)
		}
		if tc.width > 0 && lipgloss.Width(got) > tc.width {
			t.Errorf("truncateTail(%q, %d) width %d exceeds budget", tc.in, tc.width, lipgloss.Width(got))
		}
	}
}

// TestValidateNameRejectsControlChars verifies the create path refuses names
// containing control characters, so interactive creation cannot introduce a
// file whose name would later need escaping on display (defense in depth).
func TestValidateNameRejectsControlChars(t *testing.T) {
	bad := []string{"a\x1bb", "a\nb", "a\tb", "a\x00b", "a\x07b", "a\x7fb"}
	for _, name := range bad {
		if err := validateName(name); err == nil {
			t.Errorf("validateName(%q) = nil, want error", name)
		}
	}
	good := []string{"main.go", "my-file.txt", "café.md", "a b c"}
	for _, name := range good {
		if err := validateName(name); err != nil {
			t.Errorf("validateName(%q) = %v, want nil", name, err)
		}
	}
}
