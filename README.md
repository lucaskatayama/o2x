# o2x

OAuth2 CLI with pluggable flows.

## Install

```bash
brew tap lucaskatayama/tap
brew install lucaskatayama/tap/o2x
```

## Quick Start

```bash
export OAUTH2_AUTH_URL=https://your-provider.com/authorize
export OAUTH2_TOKEN_URL=https://your-provider.com/oauth/token
export OAUTH2_CLIENT_ID=your-client-id
export OAUTH2_CLIENT_SECRET=your-client-secret

o2x authorize
o2x token -n  # print access token
```

## Environment Variables

| Variable | Description |
| --- | --- |
| `OAUTH2_AUTH_URL` | Authorization URL |
| `OAUTH2_TOKEN_URL` | Token URL |
| `OAUTH2_CLIENT_ID` | Client ID |
| `OAUTH2_CLIENT_SECRET` | Client Secret |
| `OAUTH2_SCOPE` | Scopes (default: openid profile email) |
| `OAUTH2_JWKS_URI` | JWKS URI for token verification |

## Commands

| Command | Description |
| --- | --- |
| `authorize` | Start OAuth2 flow |
| `token` | Print access token (`-n` for no-newline) |
| `decode` | Decode JWT (`-t id-token` or `access-token`) |
| `verify` | Verify token signature |
| `refresh` | Refresh token |
| `revoke` | Revoke token |
| `introspect` | Introspect token |
| `userinfo` | Get user info |

## Flows

- `authorization_code` (default)
- `client_credentials`

Use `-f clientcreds` to switch.