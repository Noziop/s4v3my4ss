package backup

import (
	"strings"
	"testing"
)

func TestGenerateBackupID(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		inputName     string
		expectedParts int
	}{
		{
			name:          "Simple name",
			inputName:     "test",
			expectedParts: 3, // name_date_hash
		},
		{
			name:          "Name with spaces",
			inputName:     "my backup",
			expectedParts: 3,
		},
		{
			name:          "Name with special characters",
			inputName:     "backup!@#$%^&*()",
			expectedParts: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			result := generateBackupID(tc.inputName)

			// Check that the result is not empty
			if result == "" {
				t.Errorf("generateBackupID(%s) returned empty string", tc.inputName)
			}

			// Check that the ID has the expected format (name_date_hash)
			parts := strings.Split(result, "_")
			if len(parts) != tc.expectedParts {
				t.Errorf("generateBackupID(%s) = %s, expected %d parts but got %d", 
					tc.inputName, result, tc.expectedParts, len(parts))
			}

			// Check that special characters are properly sanitized
			if strings.ContainsAny(parts[0], "!@#$%^&*()") {
				t.Errorf("generateBackupID(%s) = %s, name part contains special characters", 
					tc.inputName, result)
			}
		})
	}
}