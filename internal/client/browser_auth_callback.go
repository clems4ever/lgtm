package client

import (
	"fmt"
	"os/exec"
	"runtime"
)

// BrowserAuthCallback facilitates browser-based authentication by opening a URL
// and waiting for the user to provide the authentication code.
type BrowserAuthCallback struct {
	codeCh chan string // Channel to receive the authentication code
}

// NewBrowserAuthCallback creates a new instance of BrowserAuthCallback.
func NewBrowserAuthCallback() *BrowserAuthCallback {
	return &BrowserAuthCallback{
		codeCh: make(chan string),
	}
}

// Receive sends the authentication code to the callback's channel.
func (c *BrowserAuthCallback) Receive(code string) {
	c.codeCh <- code
}

// Wait opens the browser with the provided URL and waits for the user to enter the authentication code.
// If the browser fails to open, the user is prompted to manually visit the URL.
func (c *BrowserAuthCallback) Wait(url string) (string, error) {
	fmt.Println("Opening browser for GitHub login...")
	fmt.Printf("If the browser does not open, visit %s\n", url)
	switch runtime.GOOS {
	case "darwin":
		// macOS: Use the "open" command to open the URL in the default browser
		exec.Command("open", url).Start()
	default: // linux and others
		// Linux: Use the "xdg-open" command to open the URL in the default browser
		exec.Command("xdg-open", url).Start()
	}
	// Prompt the user to manually enter the code if needed
	go func() {
		userCode := ""
		fmt.Print("and enter the code in the URL: ")
		fmt.Scanln(&userCode)
		c.codeCh <- userCode
	}()
	// Wait for the code to be received
	code := <-c.codeCh
	return code, nil
}
