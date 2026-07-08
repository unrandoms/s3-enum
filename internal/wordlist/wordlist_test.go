package wordlist

import (
	"strings"
	"testing"
)

func TestStripTLD(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"acme", "acme"},
		{"acme.com", "acme"},
		{"sub.acme.com", "sub.acme"},
		{"ACME.COM", "acme"},
	}
	for _, c := range cases {
		got := stripTLD(c.in)
		if got != c.want {
			t.Errorf("stripTLD(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestGenerateContainsBase(t *testing.T) {
	candidates := Generate("acme.com", nil, false)
	found := false
	for _, c := range candidates {
		if c == "acme" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Generate did not include bare base name 'acme'")
	}
}

func TestGenerateContainsSuffixVariants(t *testing.T) {
	candidates := Generate("acme", nil, false)
	set := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		set[c] = true
	}
	required := []string{"acme-dev", "acme-staging", "acme-prod", "acme-backup", "acme-logs"}
	for _, r := range required {
		if !set[r] {
			t.Errorf("Generate missing expected candidate %q", r)
		}
	}
}

func TestGeneratePermutations(t *testing.T) {
	custom := []string{"myword"}
	candidates := Generate("acme", custom, true)
	set := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		set[c] = true
	}
	if !set["acme-myword"] {
		t.Error("permutations: missing acme-myword")
	}
	if !set["myword-acme"] {
		t.Error("permutations: missing myword-acme")
	}
}

func TestGenerateNoDuplicates(t *testing.T) {
	candidates := Generate("acme.com", nil, true)
	seen := make(map[string]int)
	for _, c := range candidates {
		seen[c]++
	}
	for k, v := range seen {
		if v > 1 {
			t.Errorf("duplicate candidate %q (count %d)", k, v)
		}
	}
}

func TestGenerateAllLowercase(t *testing.T) {
	candidates := Generate("ACME.COM", nil, false)
	for _, c := range candidates {
		if c != strings.ToLower(c) {
			t.Errorf("candidate not lowercase: %q", c)
		}
	}
}
