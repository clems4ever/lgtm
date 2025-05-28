package client

import (
	"fmt"
	"net/http"

	"github.com/clems4ever/lgtm/internal/github"
)

func HandlerCallback(receiver github.AuthCodeReceiver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, `
			<html>
			<body>
				<p>You can close this window now.</p>
				<script>
					window.close();
				</script>
			</body>
			</html>
		`)
		receiver.Receive(code)
	}
}
