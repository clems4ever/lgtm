# LGTM - Lightweight GitHub Team Manager

LGTM is a lightweight tool designed to streamline the approval process for GitHub pull requests. It consists of a server and a client that work together to forward pull requests to available approvers in your team.

## Features

- **WebSocket Communication**: The server and client communicate via WebSocket for real-time updates.
- **GitHub Integration**: Authenticate with GitHub and interact with repositories and pull requests.
- **Approval Forwarding**: Automatically forward pull requests to available approvers.
- **Browser-Based Authentication**: Authenticate using OAuth2 with GitHub via a browser.

## Getting Started

### Prerequisites

- A GitHub account with access to the repositories you want to manage.

### Starting the Client

The client connects to the server, authenticates with GitHub, and listens for pull request approval requests. Pull requests are submitted via a tiny web UI served by the client at the address you specify.

1. Run the client:
   ```bash
   go run ./internal/client/cmd.go client --server-url "ws://lgtm.clems4ever.com/ws" --addr ":8081" --auth-token "your-shared-token"
   ```

   - `--server-url`: The WebSocket URL of the server.
   - `--addr`: The address and port the local client server will listen on to serve the UI (default is `:8081`).
   - `--auth-token`: The shared token provided by the server owner and used to authenticate with the server.

2. The client will start and open a browser for GitHub authentication. Follow the instructions to log in.

3. Once authenticated, visit the local client server's home page:
   ```
   http://127.0.0.1:8081/
   ```

   Use the form to submit pull request links for approval.

### Starting the Server (only for admins)

The server listens for WebSocket connections from clients and forwards pull requests to approvers.

1. Run the server:
   ```bash
   go run ./internal/server/cmd.go server --addr ":8080" --auth-token "your-shared-token"
   ```

   - `--addr`: The address and port the server will listen on (default is `:8080`).
   - `--auth-token`: A shared token used to authenticate clients.

2. The server will start and log the listening address:
   ```
   Server listening on :8080
   ```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests to improve the project.

### Build the Project

1. Clone the repository:
   ```bash
   git clone https://github.com/clems4ever/lgtm.git
   cd lgtm
   ```

2. Build the project:
   ```bash
   go build ./...
   ```

### Run the Client and Server

You can run the command-line tool without compiling. Just do:

   ```bash
   go run main.go
   ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.