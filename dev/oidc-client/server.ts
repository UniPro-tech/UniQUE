const issuer = process.env.OIDC_ISSUER ?? "http://localhost:8000";
const issuerInternal = process.env.OIDC_ISSUER_INTERNAL ?? issuer;
const clientId = process.env.OIDC_CLIENT_ID ?? "01J00000000000000000000001";
const redirectUri = process.env.OIDC_REDIRECT_URI ?? "http://localhost:3002/callback";
const port = Number(process.env.PORT ?? 3002);

type LoginState = {
  state: string;
  nonce: string;
  verifier: string;
  expiresAt: number;
};

const loginStates = new Map<string, LoginState>();
const encoder = new TextEncoder();

const base64Url = (input: Uint8Array) =>
  Buffer.from(input).toString("base64url");

const randomValue = () => base64Url(crypto.getRandomValues(new Uint8Array(32)));

const sha256 = async (value: string) =>
  new Uint8Array(await crypto.subtle.digest("SHA-256", encoder.encode(value)));

const cookies = (request: Request) =>
  Object.fromEntries(
    (request.headers.get("cookie") ?? "")
      .split(";")
      .map((value) => value.trim().split("=", 2))
      .filter(([key]) => key),
  );

const jsonResponse = (value: unknown, status = 200) =>
  new Response(JSON.stringify(value, null, 2), {
    status,
    headers: { "content-type": "application/json; charset=utf-8" },
  });

const page = (title: string, body: string) =>
  new Response(`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>${title}</title><style>body{font:16px system-ui;max-width:900px;margin:3rem auto;padding:0 1rem}a{display:inline-block;padding:.7rem 1rem;background:#155eef;color:white;border-radius:.4rem;text-decoration:none}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:#f4f4f5;padding:1rem}</style></head><body><h1>${title}</h1>${body}</body></html>`, { headers: { "content-type": "text/html; charset=utf-8" } });

const decodeJwtPart = (value: string) => JSON.parse(Buffer.from(value, "base64url").toString());

async function verifyIdToken(token: string, nonce: string) {
  const [encodedHeader, encodedClaims, encodedSignature] = token.split(".");
  if (!encodedHeader || !encodedClaims || !encodedSignature) throw new Error("ID token is malformed");

  const header = decodeJwtPart(encodedHeader) as { kid?: string; alg?: string };
  const claims = decodeJwtPart(encodedClaims) as { iss?: string; aud?: string | string[]; nonce?: string; exp?: number };
  const jwks = (await (await fetch(`${issuerInternal}/.well-known/jwks.json`)).json()) as { keys: JsonWebKey[] };
  const key = jwks.keys.find((candidate) => candidate.kid === header.kid);
  if (!key || header.alg !== "RS256") throw new Error("matching RS256 JWKS key was not found");

  const cryptoKey = await crypto.subtle.importKey(
    "jwk",
    key,
    { name: "RSASSA-PKCS1-v1_5", hash: "SHA-256" },
    false,
    ["verify"],
  );
  const validSignature = await crypto.subtle.verify(
    "RSASSA-PKCS1-v1_5",
    cryptoKey,
    Buffer.from(encodedSignature, "base64url"),
    encoder.encode(`${encodedHeader}.${encodedClaims}`),
  );
  const audience = Array.isArray(claims.aud) ? claims.aud : [claims.aud];
  if (!validSignature || claims.iss !== issuer || !audience.includes(clientId) || claims.nonce !== nonce || !claims.exp || claims.exp * 1000 < Date.now()) {
    throw new Error("ID token validation failed");
  }
  return claims;
}

Bun.serve({
  port,
  async fetch(request) {
    const url = new URL(request.url);

    if (url.pathname === "/health") return jsonResponse({ status: "ok" });

    if (url.pathname === "/login") {
      const sessionId = randomValue();
      const verifier = randomValue();
      const login = { state: randomValue(), nonce: randomValue(), verifier, expiresAt: Date.now() + 10 * 60 * 1000 };
      loginStates.set(sessionId, login);
      const parameters = new URLSearchParams({
        client_id: clientId,
        redirect_uri: redirectUri,
        response_type: "code",
        scope: "openid profile email",
        state: login.state,
        nonce: login.nonce,
        code_challenge: base64Url(await sha256(verifier)),
        code_challenge_method: "S256",
      });
      return new Response(null, {
        status: 302,
        headers: {
          location: `${issuer}/authorization?${parameters}`,
          "set-cookie": `oidc_test_session=${sessionId}; HttpOnly; SameSite=Lax; Path=/`,
        },
      });
    }

    if (url.pathname === "/callback") {
      const sessionId = cookies(request).oidc_test_session;
      const login = sessionId ? loginStates.get(sessionId) : undefined;
      loginStates.delete(sessionId ?? "");
      if (!login || login.expiresAt < Date.now() || url.searchParams.get("state") !== login.state) return page("OIDC test failed", "<p>Missing, expired, or invalid state.</p>");

      const code = url.searchParams.get("code");
      if (!code) return page("OIDC test failed", `<pre>${url.searchParams.get("error") ?? "No authorization code returned"}</pre>`);

      const tokenResponse = await fetch(`${issuerInternal}/token`, {
        method: "POST",
        headers: { "content-type": "application/x-www-form-urlencoded" },
        body: new URLSearchParams({ grant_type: "authorization_code", client_id: clientId, code, redirect_uri: redirectUri, code_verifier: login.verifier }),
      });
      const tokens = await tokenResponse.json() as { access_token?: string; id_token?: string; [key: string]: unknown };
      if (!tokenResponse.ok || !tokens.id_token || !tokens.access_token) return page("OIDC test failed", `<pre>${escapeHtml(JSON.stringify(tokens, null, 2))}</pre>`);

      try {
        const [idToken, userInfo] = await Promise.all([
          verifyIdToken(tokens.id_token, login.nonce),
          fetch(`${issuerInternal}/userinfo`, { headers: { authorization: `Bearer ${tokens.access_token}` } }).then((response) => response.json()),
        ]);
        return page("OIDC test succeeded", `<p>Authorization Code Flow with PKCE completed.</p><h2>ID Token</h2><pre>${escapeHtml(JSON.stringify(idToken, null, 2))}</pre><h2>UserInfo</h2><pre>${escapeHtml(JSON.stringify(userInfo, null, 2))}</pre>`);
      } catch (error) {
        return page("OIDC test failed", `<pre>${escapeHtml(error instanceof Error ? error.message : String(error))}</pre>`);
      }
    }

    return page("UniQUE OIDC Test Client", `<p>Issuer: ${escapeHtml(issuer)}</p><p>Client ID: ${escapeHtml(clientId)}</p><a href="/login">Start OIDC login</a>`);
  },
});

function escapeHtml(value: string) {
  return value.replace(/[&<>"']/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character] as string);
}
