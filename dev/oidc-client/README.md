# UniQUE OIDC Test Client

This local-only Hono relying party verifies the Authorization Code Flow with
PKCE, ID Token signature and claims, UserInfo, Refresh Token rotation, and RFC
7009 revocation.

Copy `.env.example` to `.env`, then start the complete test stack from the
repository root:

```sh
docker compose --profile oidc up -d --build
```

Sign in with the development user `test` / `testpassword`, then approve the
consent request.

The client ID and redirect URI are seeded by `db-migrate` only when
`SEED_TEST_USER=true`. The client is a Public Client and therefore uses PKCE
without a client secret.

Run the automated protocol smoke test after the stack is ready:

```sh
bun --cwd dev/oidc-client run test:oidc
```

The smoke test performs the same Authorization Code Flow with PKCE using the
development user, validates the returned ID Token in this client, calls
UserInfo, verifies that Refresh Token rotation invalidates the old token set,
and verifies RFC 7009 revocation. It intentionally bypasses browser rendering
of the consent page; use the browser flow to test the Frontend UI itself.

## Structure

- `server.ts`: Bun server entrypoint
- `src/app.tsx`: Hono route handlers and flow orchestration
- `src/oidc.ts`: Authorization, Token, UserInfo, and Revocation requests
- `src/session.ts`: transient browser-session state and PKCE values
- `src/ui.tsx`: Hono JSX flow visualization and expandable debug-log components
- `public/styles.css`: test-client-only page styles
