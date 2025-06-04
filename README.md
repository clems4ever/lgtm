# LGTM - Auto Approve GitHub Pull Requests

LGTM is a simple project that automatically approves GitHub pull requests. It consists of a server and a client that work together to forward pull requests to available approvers in your team.

## Features

- **WebSocket Communication**: The server and client communicate via WebSocket for real-time updates.
- **GitHub Integration**: Authenticate with GitHub and interact with repositories and pull requests.
- **Approval Forwarding**: Automatically forward pull requests to available approvers.

## Getting Started

### Quick Start with Docker

The easiest way to deploy LGTM is using Docker Compose.  
See [examples/docker/README.md](examples/docker/README.md) for instructions and environment variable requirements.

---

### Prerequisites

- A GitHub account with access to the repositories you want to manage.
- A **GitHub Classic Personal Access Token** with `repo` and `read:user` permissions for the client.

### Starting the Client

The client connects to the server, authenticates with GitHub, and listens for pull request approval requests. Pull requests are submitted via a tiny web UI served by the client at the address you specify.

**You must provide your GitHub Classic Personal Access Token via the `LGTM_GITHUB_TOKEN` environment variable.**  
**You must also provide the shared authentication token via the `LGTM_API_AUTH_TOKEN` environment variable.**

1. Run the client:
   ```bash
   export LGTM_GITHUB_TOKEN=ghp_xxx... # must have 'repo' and 'read:user' scopes
   export LGTM_API_AUTH_TOKEN=your-shared-token
   go run ./internal/client/cmd.go client
   ```

   - `--server-url`: The WebSocket URL of the server (default: `https://lgtm.clems4ever.com`).
   - `--reconnect-interval`: Time between two reconnection attempts (default: `15s`).
   - `--ping-interval`: Interval for websocket ping messages (default: `10s`).

2. The client will start and use the provided GitHub token to authenticate. If the token is missing, the client will exit with an error. At this point the client should be able to handle PR approvals automatically.

### Starting the Server (only for admins)

The server listens for WebSocket connections from clients and forwards pull requests to approvers.

**You must provide the following secrets as environment variables:**
- `LGTM_API_AUTH_TOKEN`: Shared authentication token for clients.
- `LGTM_GITHUB_CLIENT_ID`: GitHub OAuth app client ID.
- `LGTM_GITHUB_CLIENT_SECRET`: GitHub OAuth app client secret.
- `LGTM_SESSION_STORE_ENCRYPTION_KEY`: Encryption key for session cookies.

1. Run the server:
   ```bash
   export LGTM_API_AUTH_TOKEN=your-shared-token
   export LGTM_GITHUB_CLIENT_ID=your-gh-client-id
   export LGTM_GITHUB_CLIENT_SECRET=your-gh-client-secret
   export LGTM_SESSION_STORE_ENCRYPTION_KEY=your-session-key
   go run ./internal/server/cmd.go server --addr ":8080" --base-url "https://your-lgtm-url"
   ```

   - `--addr`: The address and port the server will listen on (default: `:8080`).
   - `--base-url`: The base URL of the service being served (for OAuth2 redirect).
   - `--auth-server-url`: The URL to the GitHub OAuth server (default: `https://github.com/login/oauth`).
   - `--ping-interval`: Interval for websocket ping messages (default: `10s`).

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