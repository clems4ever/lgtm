# LGTM Docker Example

## GitHub Token Requirements

When running the LGTM client in Docker, you must provide a **GitHub Classic Personal Access Token** via the `LGTM_GITHUB_TOKEN` environment variable.

- The token **must be a classic token** (not a fine-grained token).
- The token must have the following permissions (scopes):
  - `repo`
  - `read:user`

You can create a classic token in your [GitHub Developer Settings](https://github.com/settings/tokens?type=classic).

Add the token to your `client.env` file as follows:

```
LGTM_GITHUB_TOKEN=ghp_xxx...
```

**Note:**  
If the token does not have the required permissions, the client will not be able to authenticate or approve pull requests.

## Required Environment Variables

You must provide the following secrets in your environment files:

- In `client.env`:
  - `LGTM_GITHUB_TOKEN` (see above)
  - `LGTM_API_AUTH_TOKEN` (shared token, must match the server)
  - Any other client-specific configuration

- In `server.env`:
  - `LGTM_API_AUTH_TOKEN` (shared token, must match the client)
  - `LGTM_GITHUB_CLIENT_SECRET` (GitHub OAuth app client secret)
  - `LGTM_SESSION_STORE_ENCRYPTION_KEY` (random string for session encryption)
  - Any other server-specific configuration

## Docker Compose Configuration

In your `docker-compose.yml`, you must provide the GitHub OAuth app client ID to the server using the `--client-id` flag:

```yaml
services:
  server:
    # ...existing code...
    command:
      - "/app"
      - "server"
      - "--client-id"
      - "YOUR_GITHUB_CLIENT_ID"
      - "--base-url"
      - "https://lgtm.clems4ever.com"
    # ...existing code...
```

Make sure the client ID matches your GitHub OAuth app and the secrets are set in the corresponding `.env` files.
