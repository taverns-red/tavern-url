package service

import (
	"regexp"
	"testing"
)

func TestGenerateSlug_Length(t *testing.T) {
	slug, err := GenerateSlug()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slug) != 6 {
		t.Errorf("expected slug length 6, got %d: %q", len(slug), slug)
	}
}

func TestGenerateSlug_Base62(t *testing.T) {
	base62 := regexp.MustCompile(`^[0-9a-zA-Z]+$`)
	for i := 0; i < 100; i++ {
		slug, err := GenerateSlug()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if !base62.MatchString(slug) {
			t.Errorf("slug contains non-Base62 characters: %q", slug)
		}
	}
}

func TestGenerateSlug_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		slug, err := GenerateSlug()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if seen[slug] {
			t.Errorf("duplicate slug generated: %q", slug)
		}
		seen[slug] = true
	}
}

func TestGenerateSlug_ReturnsError(t *testing.T) {
	// On a healthy system, GenerateSlug should never return an error.
	// This test documents the contract: the function returns (string, error)
	// rather than panicking.
	slug, err := GenerateSlug()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug == "" {
		t.Error("expected non-empty slug on success")
	}
}

func TestValidateCustomSlug_Valid(t *testing.T) {
	valid := []string{"abc", "my-link", "Spring-Gala-2026", "a1b2c3"}
	for _, slug := range valid {
		if err := ValidateCustomSlug(slug); err != nil {
			t.Errorf("expected valid slug %q, got error: %v", slug, err)
		}
	}
}

func TestValidateCustomSlug_TooShort(t *testing.T) {
	if err := ValidateCustomSlug("ab"); err == nil {
		t.Error("expected error for 2-char slug, got nil")
	}
}

func TestValidateCustomSlug_TooLong(t *testing.T) {
	long := make([]byte, 65)
	for i := range long {
		long[i] = 'a'
	}
	if err := ValidateCustomSlug(string(long)); err == nil {
		t.Error("expected error for 65-char slug, got nil")
	}
}

func TestValidateCustomSlug_InvalidChars(t *testing.T) {
	invalid := []string{"hello world", "foo@bar", "slug/path", "emoji😀", "under_score"}
	for _, slug := range invalid {
		if err := ValidateCustomSlug(slug); err == nil {
			t.Errorf("expected error for invalid slug %q, got nil", slug)
		}
	}
}

func TestValidateCustomSlug_MaxLength(t *testing.T) {
	maxSlug := make([]byte, 64)
	for i := range maxSlug {
		maxSlug[i] = 'a'
	}
	if err := ValidateCustomSlug(string(maxSlug)); err != nil {
		t.Errorf("expected valid 64-char slug, got error: %v", err)
	}
}

func TestValidateCustomSlug_MinLength(t *testing.T) {
	if err := ValidateCustomSlug("abc"); err != nil {
		t.Errorf("expected valid 3-char slug, got error: %v", err)
	}
}
