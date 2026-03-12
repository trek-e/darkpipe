// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package logutil

import (
	"testing"
)

func TestRedactEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal address", "sender@example.com", "s***r@example.com"},
		{"three-char local", "abc@example.com", "a***c@example.com"},
		{"two-char local", "ab@example.com", "a*@example.com"},
		{"single-char local", "a@example.com", "*@example.com"},
		{"empty local", "@example.com", "@example.com"},
		{"no at sign", "no-at-sign", "no-at-sign"},
		{"empty string", "", ""},
		{"long local part", "john.doe.smith@example.com", "j***h@example.com"},
		{"multiple at signs", "user@sub@example.com", "u***b@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactEmail(tt.input)
			if got != tt.want {
				t.Errorf("RedactEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRedactQueryParams(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"emailaddress param",
			"emailaddress=user@example.com",
			"emailaddress=%5BREDACTED%5D",
		},
		{
			"token param",
			"token=abc123secret",
			"token=%5BREDACTED%5D",
		},
		{
			"both email and token",
			"email=user@example.com&token=secret123",
			"email=%5BREDACTED%5D&token=%5BREDACTED%5D",
		},
		{
			"no sensitive params",
			"format=json&page=2",
			"format=json&page=2",
		},
		{
			"empty query string",
			"",
			"",
		},
		{
			"mixed sensitive and safe",
			"emailaddress=user@example.com&format=xml",
			"emailaddress=%5BREDACTED%5D&format=xml",
		},
		{
			"case-insensitive key matching",
			"EmailAddress=user@example.com",
			"EmailAddress=%5BREDACTED%5D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactQueryParams(tt.input)
			if got != tt.want {
				t.Errorf("RedactQueryParams(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
