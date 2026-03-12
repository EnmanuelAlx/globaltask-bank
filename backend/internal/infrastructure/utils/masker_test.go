package utils

import (
	"testing"
)

func TestMaskID(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"12345678", "****5678"},
		{"1234", "****"},
		{"123", "***"},
		{"", ""},
		{"ABCDEFGH", "****EFGH"},
	}

	for _, tt := range tests {
		result := MaskID(tt.id)
		if result != tt.expected {
			t.Errorf("MaskID(%s) = %s; want %s", tt.id, result, tt.expected)
		}
	}
}

func TestMaskName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"John Doe", "J*** D**"},
		{"Alice", "A****"},
		{"B C", "B C"},
		{"", ""},
		{"  ", ""},
		{"John Von Neumann", "J*** V** N******"},
	}

	for _, tt := range tests {
		result := MaskName(tt.name)
		if result != tt.expected {
			t.Errorf("MaskName(%s) = %s; want %s", tt.name, result, tt.expected)
		}
	}
}

func TestMaskMap(t *testing.T) {
	data := map[string]interface{}{
		"application_id":    "123-456",
		"borrower_name":     "John Doe",
		"identity_document": "12345678",
		"other_field":       123.45,
		"metadata": map[string]interface{}{
			"full_name": "Alice Smith",
		},
	}

	masked := MaskMap(data)

	if masked["borrower_name"] != "J*** D**" {
		t.Errorf("Expected borrower_name to be masked, got %v", masked["borrower_name"])
	}
	if masked["identity_document"] != "****5678" {
		t.Errorf("Expected identity_document to be masked, got %v", masked["identity_document"])
	}
	if masked["application_id"] != "123-456" {
		t.Errorf("Expected application_id NOT to be masked, got %v", masked["application_id"])
	}
	if metadata, ok := masked["metadata"].(map[string]interface{}); ok {
		if metadata["full_name"] != "A**** S****" {
			t.Errorf("Expected nested full_name to be masked, got %v", metadata["full_name"])
		}
	} else {
		t.Error("Expected nested metadata map")
	}
}
