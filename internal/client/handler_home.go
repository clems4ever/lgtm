package client

import (
	"html/template"
	"net/http"
)

var homeTemplate = template.Must(template.New("home").Parse(`
<html>
<body>
    <h1>Submit a GitHub Pull Request Link</h1>
    <p>The Pull Request will be forwarded to an available approver if there is any.</p>
    <form id="approve-form">
        <label>PR Link: <input type="text" name="pr_link" id="pr_link" size="80"/></label>
        <input type="submit" value="Approve"/>
    </form>
    <div id="result"></div>
    <h2>Available Approvers</h2>
    <ul>
    {{range .Approvers}}
        <li>{{.}}</li>
    {{else}}
        <li><em>No approvers available</em></li>
    {{end}}
    </ul>
    <script>
    document.getElementById('approve-form').onsubmit = async function(e) {
        e.preventDefault();
        const prInput = document.getElementById('pr_link');
        const prLink = prInput.value;
        await fetch('/approve', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: 'pr_link=' + encodeURIComponent(prLink)
        });
        document.getElementById('result').innerText = "\u2705 PR submitted for approval!";
        prInput.value = '';
    };
    </script>
</body>
</html>
`))

// HomeTemplateArgs represents the data passed to the home page template.
// It includes a list of available approvers to display on the page.
type HomeTemplateArgs struct {
	Approvers []string // List of available approvers
}

// handlerHome serves the home page of the client.
// It displays a form for submitting GitHub PR links and a list of available approvers.
//
// Behavior:
// - For GET requests, it renders the home page with the list of approvers.
// - For other HTTP methods, it responds with a 404 Not Found.
//
// Parameters:
// - w: The HTTP response writer.
// - r: The HTTP request.
func (c *Client) handlerHome(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.availableApproversMu.Lock()
		defer c.availableApproversMu.Unlock()
		homeTemplate.Execute(w, HomeTemplateArgs{
			Approvers: c.availableApprovers,
		})
		return
	}
	http.NotFound(w, r)
}
