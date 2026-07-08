package wordlist

import (
	"bufio"
	"os"
	"strings"
)

// Generate returns all bucket name candidates derived from target.
// target may be a bare name ("acme") or a domain ("acme.com").
// When customWords is non-nil they are appended to the builtin list.
// When permutations is true, target-WORD and WORD-target variants are
// generated for every word in the combined list.
func Generate(target string, customWords []string, permutations bool) []string {
	base := stripTLD(target)

	words := make([]string, 0, len(BuiltinSuffixes)+len(customWords))
	words = append(words, BuiltinSuffixes...)
	words = append(words, customWords...)

	seen := make(map[string]struct{})
	add := func(s string) {
		s = strings.ToLower(s)
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
		}
	}

	// Plain target + raw domain variants
	add(base)
	if base != target {
		add(target)
		add("www." + target)
	}

	// target-SUFFIX and PREFIX-target from builtins
	for _, s := range BuiltinSuffixes {
		add(base + "-" + s)
		add(base + "." + s)
	}
	for _, p := range BuiltinPrefixes {
		add(p + "-" + base)
	}

	// permutation pass over combined wordlist
	if permutations {
		for _, w := range words {
			add(base + "-" + w)
			add(w + "-" + base)
		}
	}

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

// stripTLD removes the rightmost dot-separated label from a domain name.
// "acme.com" → "acme", "sub.acme.co.uk" → "sub.acme.co", bare "acme" → "acme".
func stripTLD(s string) string {
	s = strings.ToLower(s)
	parts := strings.Split(s, ".")
	if len(parts) <= 1 {
		return s
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

// LoadFile reads one word per line from path and returns the non-empty,
// trimmed, lowercased entries.
func LoadFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var words []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.TrimSpace(strings.ToLower(sc.Text()))
		if w != "" && !strings.HasPrefix(w, "#") {
			words = append(words, w)
		}
	}
	return words, sc.Err()
}
