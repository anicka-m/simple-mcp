package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestSearchResources(t *testing.T) {
	resourceMap := map[string]ResourceItem{
		"test://apple": {
			URI:         "test://apple",
			Description: "A red fruit",
			Content:     "Static content about apples.",
		},
		"test://banana": {
			URI:         "test://banana",
			Description: "A yellow fruit",
			Content:     "Static content about bananas.",
		},
		"test://cherry": {
			URI:         "test://cherry",
			Description: "A small red fruit",
			Content:     "Static content about cherries.",
		},
	}

	tests := []struct {
		name          string
		query         string
		expectedFound int
		contains      []string
		notContains   []string
		expectError   bool
	}{
		{
			name:          "Search by URI",
			query:         "apple",
			expectedFound: 1,
			contains:      []string{"test://apple"},
			notContains:   []string{"test://banana", "test://cherry"},
		},
		{
			name:          "Search by Description",
			query:         "yellow",
			expectedFound: 1,
			contains:      []string{"test://banana"},
		},
		{
			name:          "Search by Content",
			query:         "cherries",
			expectedFound: 1,
			contains:      []string{"test://cherry"},
		},
		{
			name:          "Search multiple",
			query:         "red",
			expectedFound: 2,
			contains:      []string{"test://apple", "test://cherry"},
			notContains:   []string{"test://banana"},
		},
		{
			name:          "Complex Regex (alternation)",
			query:         "apple|banana",
			expectedFound: 2,
			contains:      []string{"test://apple", "test://banana"},
			notContains:   []string{"test://cherry"},
		},
		{
			name:          "Complex Regex (character class and wildcards)",
			query:         "r[e-i]d",
			expectedFound: 2,
			contains:      []string{"test://apple", "test://cherry"},
		},
		{
			name:          "No results",
			query:         "durian",
			expectedFound: 0,
			contains:      []string{"No resources matched"},
		},
		{
			name:        "Invalid regex",
			query:       "[",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := searchResources(resourceMap, tt.query)
			if (err != nil) != tt.expectError {
				t.Errorf("searchResources() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if tt.expectError {
				return
			}

			if tt.expectedFound > 0 {
				expectedFoundStr := fmt.Sprintf("Found %d matching resources", tt.expectedFound)
				if !strings.Contains(result, expectedFoundStr) {
					t.Errorf("result expected to contain %q, but didn't. Result: %q", expectedFoundStr, result)
				}
			}

			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("result expected to contain %q, but didn't. Result: %q", c, result)
				}
			}
			for _, nc := range tt.notContains {
				if strings.Contains(result, nc) {
					t.Errorf("result expected NOT to contain %q, but did. Result: %q", nc, result)
				}
			}
		})
	}
}
