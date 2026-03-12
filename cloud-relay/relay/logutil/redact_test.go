// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package logutil

import (
	"reflect"
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

func TestRedactEmails(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			"multiple addresses",
			[]string{"sender@example.com", "a@test.org", "ab@foo.net"},
			[]string{"s***r@example.com", "*@test.org", "a*@foo.net"},
		},
		{"empty slice", []string{}, []string{}},
		{"nil slice", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactEmails(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedactEmails(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
