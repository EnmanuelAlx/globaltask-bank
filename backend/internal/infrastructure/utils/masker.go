package utils

import (
	"strings"
)

// MaskID masks an ID leaving only the last 4 characters visible.
// Example: "12345678" -> "****5678"
func MaskID(id string) string {
	if len(id) <= 4 {
		return strings.Repeat("*", len(id))
	}
	return strings.Repeat("*", len(id)-4) + id[len(id)-4:]
}

// MaskName masks a name leaving only the first character of each word.
// Example: "John Doe" -> "J*** D**"
func MaskName(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return ""
	}
	maskedWords := make([]string, len(words))
	for i, word := range words {
		if len(word) <= 1 {
			maskedWords[i] = word
			continue
		}
		maskedWords[i] = string(word[0]) + strings.Repeat("*", len(word)-1)
	}
	return strings.Join(maskedWords, " ")
}

// MaskMap returns a copy of the map with sensitive fields masked.
// It handles common PII keys like identity_document and borrower_name.
func MaskMap(data map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for k, v := range data {
		lowerK := strings.ToLower(k)
		switch {
		case strings.Contains(lowerK, "identity_document") || strings.Contains(lowerK, "document_number") || lowerK == "doc":
			if s, ok := v.(string); ok {
				masked[k] = MaskID(s)
			} else {
				masked[k] = v
			}
		case strings.Contains(lowerK, "full_name") || strings.Contains(lowerK, "borrower_name") || lowerK == "name":
			if s, ok := v.(string); ok {
				masked[k] = MaskName(s)
			} else {
				masked[k] = v
			}
		case lowerK == "email":
			// Simple email masking: keep first letter and domain
			if s, ok := v.(string); ok {
				parts := strings.Split(s, "@")
				if len(parts) == 2 && len(parts[0]) > 1 {
					masked[k] = string(parts[0][0]) + "***@" + parts[1]
				} else {
					masked[k] = "***@***"
				}
			} else {
				masked[k] = v
			}
		default:
			if subMap, ok := v.(map[string]interface{}); ok {
				masked[k] = MaskMap(subMap)
			} else if subMap, ok := v.(entityJSONB); ok {
				// Special handling for local JSONB type if we can access it
				masked[k] = MaskMap(map[string]interface{}(subMap))
			} else {
				masked[k] = v
			}
		}
	}
	return masked
}

type entityJSONB map[string]interface{}
