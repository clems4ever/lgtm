package github

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
)

func GetTokenFilePath() (string, error) {
	u, _ := user.Current()
	path := filepath.Join(u.HomeDir, ".lgtm")
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return "", fmt.Errorf("failed to mkdir: %w", err)
	}
	return filepath.Join(path, "token.json"), nil
}

func SaveToken(path string, token *oauth2.Token) error {
	// Create the file with permission 0600
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}
