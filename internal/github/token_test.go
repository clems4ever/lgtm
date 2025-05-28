package github

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestSaveAndLoadToken(t *testing.T) {
	// Use a temporary directory for token file
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")

	origToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "bearer",
		RefreshToken: "test-refresh-token",
	}

	SaveToken(tokenFile, origToken)

	loadedToken, err := LoadToken(tokenFile)
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	if loadedToken.AccessToken != origToken.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loadedToken.AccessToken, origToken.AccessToken)
	}
	if loadedToken.TokenType != origToken.TokenType {
		t.Errorf("TokenType mismatch: got %q, want %q", loadedToken.TokenType, origToken.TokenType)
	}
	if loadedToken.RefreshToken != origToken.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", loadedToken.RefreshToken, origToken.RefreshToken)
	}

	// Clean up
	os.Remove(tokenFile)
}
