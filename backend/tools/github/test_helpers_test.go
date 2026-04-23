package githubtools

import (
	"encoding/json"
	"testing"
)

func decodeGitHubToolOutput(t *testing.T, output string) map[string]interface{} {
	t.Helper()

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decode github tool output: %v\n%s", err, output)
	}
	return decoded
}
