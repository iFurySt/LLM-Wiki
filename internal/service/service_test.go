package service

import "testing"

func TestNormalizeVisibility(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":           "private",
		"tenant":     "tenant",
		"TEAM":       "team",
		"restricted": "restricted",
		"weird":      "private",
	}

	for input, want := range cases {
		if got := normalizeVisibility(input); got != want {
			t.Fatalf("normalizeVisibility(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeKey(t *testing.T) {
	t.Parallel()

	got := normalizeKey(" Product Notes ")
	if got != "product-notes" {
		t.Fatalf("normalizeKey mismatch: %q", got)
	}
}

func TestValidateKeyLike(t *testing.T) {
	t.Parallel()

	if err := validateKeyLike("slug", "valid-key_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateKeyLike("slug", "Bad Key!"); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestNormalizeAuthorType(t *testing.T) {
	t.Parallel()

	if got := normalizeAuthorType("USER"); got != "user" {
		t.Fatalf("normalizeAuthorType user mismatch: %q", got)
	}
	if got := normalizeAuthorType(""); got != "agent" {
		t.Fatalf("normalizeAuthorType default mismatch: %q", got)
	}
}
