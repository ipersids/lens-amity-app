package auth

import (
	"strings"
	"testing"
)

func TestNormDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trims surrounding whitespace",
			in:   "  Alice Example  ",
			want: "Alice Example",
		},
		{
			name: "normalizes to NFC",
			in:   " Cafe\u0301 ",
			want: "Caf\u00e9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normDisplay(tt.in)
			if got != tt.want {
				t.Fatalf("normDisplay(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trims and folds case",
			in:   "  Alice_123  ",
			want: "alice_123",
		},
		{
			name: "normalizes width before folding case",
			in:   "\uff21\uff22\uff23_123",
			want: "abc_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normKey(tt.in)
			if got != tt.want {
				t.Fatalf("normKey(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateUsernameKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name:    "minimum length",
			in:      "abc",
			wantErr: false,
		},
		{
			name:    "maximum length",
			in:      strings.Repeat("a", 32),
			wantErr: false,
		},
		{
			name:    "allows underscore and hyphen",
			in:      "alice_bob-123",
			wantErr: false,
		},
		{
			name:    "too short",
			in:      "ab",
			wantErr: true,
		},
		{
			name:    "too long",
			in:      strings.Repeat("a", 33),
			wantErr: true,
		},
		{
			name:    "rejects spaces",
			in:      "alice bob",
			wantErr: true,
		},
		{
			name:    "rejects dots",
			in:      "alice.bob",
			wantErr: true,
		},
		{
			name:    "rejects reserved usernames",
			in:      "admin",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateUsernameKey(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateUsernameKey(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name:    "minimum length",
			in:      strings.Repeat("a", 15),
			wantErr: false,
		},
		{
			name:    "maximum length",
			in:      strings.Repeat("a", 128),
			wantErr: false,
		},
		{
			name:    "too short",
			in:      strings.Repeat("a", 14),
			wantErr: true,
		},
		{
			name:    "too long",
			in:      strings.Repeat("a", 129),
			wantErr: true,
		},
		{
			name:    "counts runes instead of bytes",
			in:      strings.Repeat("\u00e4", 15),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePasswordLength(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePasswordLength(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		password   string
		userInputs []string
		wantErr    bool
	}{
		{
			name:       "accepts strong printable password",
			password:   "correct horse battery staple 2026!",
			userInputs: []string{"alice", "Alice Example"},
			wantErr:    false,
		},
		{
			name:       "rejects weak password",
			password:   "passwordpassword",
			userInputs: []string{"alice", "Alice Example"},
			wantErr:    true,
		},
		{
			name:       "rejects non-printable characters",
			password:   "correct horse battery staple 2026!\x00",
			userInputs: []string{"alice", "Alice Example"},
			wantErr:    true,
		},
		{
			name:       "rejects passwords containing user input",
			password:   "alice alice alice alice 2026!",
			userInputs: []string{"alice", "Alice Example"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePassword(tt.password, tt.userInputs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validatePassword(%q, %v) error = %v, wantErr %v", tt.password, tt.userInputs, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNameLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name:    "minimum length",
			in:      "abc",
			wantErr: false,
		},
		{
			name:    "maximum length",
			in:      strings.Repeat("a", 32),
			wantErr: false,
		},
		{
			name:    "too short",
			in:      "ab",
			wantErr: true,
		},
		{
			name:    "too long",
			in:      strings.Repeat("a", 33),
			wantErr: true,
		},
		{
			name:    "counts runes instead of bytes",
			in:      strings.Repeat("\u00e4", 3),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateNameLength(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateNameLength(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
		})
	}
}
