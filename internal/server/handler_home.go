package server

import (
	"html/template"
	"log"
	"net/http"

	_ "embed"
)

// HomeTemplateArgs represents the data passed to the home page template.
// User: the authenticated user's GitHub username.
// Approvers: the number of available approvers.
type HomeTemplateArgs struct {
	User      string // Username of the authenticated user
	Approvers int    // Number of available approvers
}

// Embed the home.html template file for rendering the home page.
//
//go:embed ui/home.html
var homeHTML string

// homeTemplate is the parsed HTML template for the home page.
var homeTemplate = template.Must(template.New("home").Parse(homeHTML))

// handlerHome serves the home page of the client.
// - For GET requests, it renders the home page with the username and number of approvers.
// - For other HTTP methods, it responds with a 404 Not Found.
func (s *Server) handlerHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	// Retrieve the authenticated username from the request context.
	username := r.Context().Value("username").(string)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the home page template with the username and approver count.
	err := homeTemplate.Execute(w, HomeTemplateArgs{
		User:      username,
		Approvers: len(s.approvalEngine.GetApprovers()),
	})
	if err != nil {
		log.Println("failed to execute template", err)
		return
	}
}
