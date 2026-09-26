const issuer = process.env.OIDC_ISSUER ?? "http://localhost:8000";
const issuerInternal = process.env.OIDC_ISSUER_INTERNAL ?? issuer;
const clientId = process.env.OIDC_CLIENT_ID ?? "01J00000000000000000000001";
const redirectUri = process.env.OIDC_REDIRECT_URI ?? "http://localhost:3002/callback";
const port = Number(process.env.PORT ?? 3002);

type TokenSet = { access_token: string; refresh_token: string; id_token: string };
type Milestone = { title: string; details: unknown };
type LoginState = {
  state: string;
  nonce: string;
  verifier: string;
  expiresAt: number;
  stage: number;
  path?: "rotated" | "direct";
  prompt?: "consent" | "none";
  events: Milestone[];
  tokens?: TokenSet;
  error?: string;
};

const loginStates = new Map<string, LoginState>();
const encoder = new TextEncoder();
const base64Url = (input: Uint8Array) => Buffer.from(input).toString("base64url");
const randomValue = () => base64Url(crypto.getRandomValues(new Uint8Array(32)));
const sha256 = async (value: string) => new Uint8Array(await crypto.subtle.digest("SHA-256", encoder.encode(value)));

const cookies = (request: Request) => Object.fromEntries((request.headers.get("cookie") ?? "").split(";").map((value) => value.trim().split("=", 2)).filter(([key]) => key));
const cookieHeader = (sessionId: string) => `oidc_test_session=${sessionId}; HttpOnly; SameSite=Lax; Path=/`;
const escapeHtml = (value: string) => value.replace(/[&<>"']/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[character] as string);
const shortToken = (token: string) => `${token.slice(0, 16)}...${token.slice(-12)}`;

function flowChart(stage: number, path?: LoginState["path"]) {
  const common = ["PKCE request", "Sign in & consent", "Code exchange", "Verify tokens"];
  const commonNodes = common.map((label, index) => `<g data-step="${index + 1}" class="${stage >= index + 1 ? "active" : ""}"><circle cx="${60 + index * 125}" cy="70" r="27"/><text x="${60 + index * 125}" y="75">${index + 1}</text><foreignObject x="${5 + index * 125}" y="104" width="110" height="24"><div>${label}</div></foreignObject></g>`).join("");
  const commonEdges = common.slice(0, -1).map((_, index) => `<path data-edge="${index + 1}" class="${stage > index + 1 ? "active" : ""}" d="M${85 + index * 125} 70H${160 + index * 125}" />`).join("");
  const rotated = path === "rotated";
  const directRevoked = stage === 6 && path === "direct";
  const rotatedRevoked = stage === 6 && rotated;
  return `<section class="flow"><svg viewBox="0 0 800 220" role="img" aria-label="OIDC flow"><text class="branch-label" x="575" y="20">Choose one path after token verification</text>${commonEdges}${commonNodes}<path class="${rotated ? "active" : ""}" d="M455 58L533 45"/><path class="${directRevoked ? "active" : ""}" d="M455 82L533 145"/><path class="${rotatedRevoked ? "active" : ""}" d="M587 45H668"/><g class="${rotated ? "active" : ""}"><circle cx="560" cy="45" r="27"/><text x="560" y="50">5A</text><foreignObject x="500" y="79" width="120" height="24"><div>Refresh rotation</div></foreignObject></g><g class="${directRevoked ? "active" : ""}"><circle cx="560" cy="145" r="27"/><text x="560" y="150">5B</text><foreignObject x="500" y="179" width="120" height="24"><div>Direct revocation</div></foreignObject></g><g class="${rotatedRevoked ? "active" : ""}"><circle cx="695" cy="45" r="27"/><text x="695" y="50">6</text><foreignObject x="635" y="79" width="120" height="24"><div>Revoke rotated set</div></foreignObject></g></svg></section>`;
}

function page(title: string, body: string, session?: LoginState, script = "") {
  const events = session?.events.length ? `<section class="milestones"><h2>Flow log</h2>${session.events.map((event, index) => `<details${index === session.events.length - 1 ? " open" : ""}><summary><span>${index + 1}</span>${escapeHtml(event.title)}</summary><pre>${escapeHtml(JSON.stringify(event.details, null, 2))}</pre></details>`).join("")}</section>` : "";
  const actions = session?.tokens && !session.error && session.stage < 6
    ? `<form method="post" action="/refresh"><button>Rotate Refresh Token</button></form><form method="post" action="/revoke"><button class="danger">Revoke Current Token Set</button></form>`
    : "";
  const sessionControls = `<label class="prompt">prompt <select id="prompt"><option value=""${!session?.prompt ? " selected" : ""}>default</option><option value="consent"${session?.prompt === "consent" ? " selected" : ""}>consent</option><option value="none"${session?.prompt === "none" ? " selected" : ""}>none</option></select></label><button id="start-login">Start OIDC login</button>${session ? `<form method="post" action="/reset"><button class="secondary">Reset test session</button></form>` : ""}`;
  const launcher = `<script>const setStage=(stage)=>{document.querySelectorAll('[data-step]').forEach((node)=>node.classList.toggle('active',Number(node.dataset.step)<=stage));document.querySelectorAll('[data-edge]').forEach((node)=>node.classList.toggle('active',Number(node.dataset.edge)<stage))};document.getElementById('start-login')?.addEventListener('click',()=>{const prompt=document.getElementById('prompt')?.value;const loginUrl=prompt?'/login?prompt='+encodeURIComponent(prompt):'/login';const popup=window.open(loginUrl,'unique-oidc-provider','popup,width=900,height=760');if(!popup){window.location.assign(loginUrl);return}setStage(2)});window.addEventListener('message',(event)=>{if(event.origin!==window.location.origin||event.data?.type!=='unique-oidc-stage')return;setStage(event.data.stage);if(event.data.stage===4)setTimeout(()=>window.location.reload(),700)});</script>`;
  return new Response(`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>${title}</title><style>body{font:16px system-ui;max-width:980px;margin:3rem auto;padding:0 1rem;color:#172033}.flow{position:sticky;top:0;margin:0 -1rem 2rem;padding:1rem;background:#fff;z-index:1}a,button,select{display:inline-block;padding:.7rem 1rem;background:#155eef;color:white;border:0;border-radius:.45rem;text-decoration:none;font:inherit;cursor:pointer}.prompt{display:inline-flex;gap:.5rem;align-items:center;margin-right:.5rem}.prompt select{background:white;color:#172033;border:1px solid #94a3b8;padding:.45rem}.secondary{background:#475569}.danger{background:#b42318}.flow svg{width:100%;overflow:visible}.flow path{stroke:#cbd5e1;stroke-width:4;fill:none}.flow path.active{stroke:#155eef;stroke-dasharray:12 8;animation:flow 1s linear infinite}.flow circle{fill:#e2e8f0}.flow g.active circle{fill:#155eef}.flow text{fill:#475569;text-anchor:middle;font-weight:700}.flow g.active text{fill:white}.flow .branch-label{font-size:12px;font-weight:400;fill:#64748b}.flow foreignObject div{text-align:center;font-size:12px;color:#475569}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:#f4f4f5;padding:1rem;border-radius:.4rem}.milestones{margin:2rem 0}.milestones details{border-top:1px solid #e2e8f0;padding:.75rem 0}.milestones summary{cursor:pointer;font-weight:600}.milestones summary span{display:inline-grid;place-items:center;width:1.5rem;height:1.5rem;margin-right:.5rem;border-radius:50%;background:#dbeafe;color:#155eef;font-size:.8rem}.milestones pre{margin:.75rem 0 0}form{display:inline-block;margin:0 .5rem .5rem 0}@keyframes flow{to{stroke-dashoffset:-20}}</style></head><body>${flowChart(session?.stage ?? 0, session?.path)}<h1>${title}</h1>${sessionControls}${body}${events}${actions}${launcher}${script}</body></html>`, { headers: { "content-type": "text/html; charset=utf-8" } });
}

function getSession(request: Request) {
  const sessionId = cookies(request).oidc_test_session;
  const session = sessionId ? loginStates.get(sessionId) : undefined;
  return { sessionId, session };
}

const decodeJwtPart = (value: string) => JSON.parse(Buffer.from(value, "base64url").toString());

async function verifyIdToken(token: string, nonce: string) {
  const [encodedHeader, encodedClaims, encodedSignature] = token.split(".");
  if (!encodedHeader || !encodedClaims || !encodedSignature) throw new Error("ID token is malformed");
  const header = decodeJwtPart(encodedHeader) as { kid?: string; alg?: string };
  const claims = decodeJwtPart(encodedClaims) as { iss?: string; aud?: string | string[]; nonce?: string; exp?: number };
  const jwks = (await (await fetch(`${issuerInternal}/.well-known/jwks.json`)).json()) as { keys: JsonWebKey[] };
  const key = jwks.keys.find((candidate) => candidate.kid === header.kid);
  if (!key || header.alg !== "RS256") throw new Error("matching RS256 JWKS key was not found");
  const cryptoKey = await crypto.subtle.importKey("jwk", key, { name: "RSASSA-PKCS1-v1_5", hash: "SHA-256" }, false, ["verify"]);
  const validSignature = await crypto.subtle.verify("RSASSA-PKCS1-v1_5", cryptoKey, Buffer.from(encodedSignature, "base64url"), encoder.encode(`${encodedHeader}.${encodedClaims}`));
  const audience = Array.isArray(claims.aud) ? claims.aud : [claims.aud];
  if (!validSignature || claims.iss !== issuer || !audience.includes(clientId) || claims.nonce !== nonce || !claims.exp || claims.exp * 1000 < Date.now()) throw new Error("ID token validation failed");
  return claims;
}

async function tokenRequest(parameters: Record<string, string>) {
  const body = new URLSearchParams(parameters);
  const request = { method: "POST", url: `${issuerInternal}/token`, headers: { "content-type": "application/x-www-form-urlencoded" }, body: Object.fromEntries(body) };
  const response = await fetch(request.url, { method: request.method, headers: request.headers, body });
  return { response, tokens: await response.json() as Partial<TokenSet> & { error?: string }, request };
}

async function userInfo(accessToken: string) {
  const request = { method: "GET", url: `${issuerInternal}/userinfo`, headers: { authorization: `Bearer ${accessToken}` } };
  const response = await fetch(request.url, { headers: request.headers });
  return { response, body: await response.json(), request };
}

Bun.serve({
  port,
  async fetch(request) {
    const url = new URL(request.url);
    if (url.pathname === "/health") return new Response(JSON.stringify({ status: "ok" }), { headers: { "content-type": "application/json" } });

    if (url.pathname === "/login") {
      const prompt = url.searchParams.get("prompt");
      if (prompt !== null && prompt !== "consent" && prompt !== "none") return page("OIDC test failed", "<p>Unsupported prompt value.</p>");
      const { sessionId: existingSessionId, session: existingSession } = getSession(request);
      // Reusing an in-progress transaction prevents a duplicate popup from replacing its state cookie.
      const reuseSession = existingSessionId && existingSession && existingSession.expiresAt >= Date.now() && existingSession.stage < 4 && !existingSession.error && existingSession.prompt === (prompt ?? undefined);
      const sessionId = reuseSession ? existingSessionId : randomValue();
      const session = reuseSession
        ? existingSession
        : { state: randomValue(), nonce: randomValue(), verifier: randomValue(), expiresAt: Date.now() + 10 * 60 * 1000, stage: 1, prompt: prompt ?? undefined, events: [] };
      if (!reuseSession) loginStates.set(sessionId, session);
      const parameters = new URLSearchParams({ client_id: clientId, redirect_uri: redirectUri, response_type: "code", scope: "openid profile email", state: session.state, nonce: session.nonce, code_challenge: base64Url(await sha256(session.verifier)), code_challenge_method: "S256" });
      if (session.prompt) parameters.set("prompt", session.prompt);
      if (!reuseSession) session.events.push({ title: "Authorization request prepared", details: { request: { method: "GET", url: `${issuer}/authorization?${parameters}`, query: Object.fromEntries(parameters) } } });
      return new Response(null, { status: 302, headers: { location: `${issuer}/authorization?${parameters}`, ...(reuseSession ? {} : { "set-cookie": cookieHeader(sessionId) }) } });
    }

    const { sessionId, session } = getSession(request);
    if (url.pathname === "/reset" && request.method === "POST") {
      if (sessionId) loginStates.delete(sessionId);
      return new Response(null, { status: 303, headers: { location: "/", "set-cookie": "oidc_test_session=; Max-Age=0; HttpOnly; SameSite=Lax; Path=/" } });
    }

    if (url.pathname === "/callback") {
      if (!session || !sessionId || session.expiresAt < Date.now() || url.searchParams.get("state") !== session.state) return page("OIDC test failed", "<p>Missing, expired, or invalid state.</p>");
      const code = url.searchParams.get("code");
      if (!code) {
        const error = url.searchParams.get("error") ?? "No authorization code returned";
        session.error = error;
        session.events.push({ title: "Authorization response returned an error", details: { parameters: Object.fromEntries(url.searchParams) } });
        return page("OIDC test failed", `<pre>${escapeHtml(error)}</pre>`, session);
      }
      const { response, tokens, request: tokenExchangeRequest } = await tokenRequest({ grant_type: "authorization_code", client_id: clientId, code, redirect_uri: redirectUri, code_verifier: session.verifier });
      if (!response.ok || !tokens.id_token || !tokens.access_token || !tokens.refresh_token) {
        session.error = "Token exchange failed";
        session.events.push({ title: "Token exchange failed", details: { request: tokenExchangeRequest, status: response.status, response: tokens } });
        return page("OIDC test failed", `<pre>${escapeHtml(JSON.stringify(tokens, null, 2))}</pre>`, session);
      }
      try {
        const [idToken, info] = await Promise.all([verifyIdToken(tokens.id_token, session.nonce), userInfo(tokens.access_token)]);
        if (!info.response.ok) throw new Error("UserInfo rejected the new access token");
        session.tokens = tokens as TokenSet;
        session.stage = 4;
        session.events.push({ title: "Authorization code exchanged and ID Token verified", details: { request: tokenExchangeRequest, status: response.status, idTokenClaims: idToken, tokenResponse: tokens } }, { title: "UserInfo accepted the access token", details: { request: info.request, status: info.response.status, response: info.body } });
        const callbackScript = `<script>if(window.opener){window.opener.postMessage({type:"unique-oidc-stage",stage:3},window.location.origin);setTimeout(()=>window.opener.postMessage({type:"unique-oidc-stage",stage:4},window.location.origin),500);setTimeout(()=>window.close(),900)}</script>`;
        return page("OIDC test succeeded", `<p>Authorization Code Flow with PKCE completed.</p><h2>ID Token</h2><pre>${escapeHtml(JSON.stringify(idToken, null, 2))}</pre><h2>UserInfo</h2><pre>${escapeHtml(JSON.stringify(info.body, null, 2))}</pre>`, session, callbackScript);
      } catch (error) {
        session.error = error instanceof Error ? error.message : String(error);
        session.events.push({ title: "Token validation failed", details: { error: session.error } });
        return page("OIDC test failed", `<pre>${escapeHtml(session.error)}</pre>`, session);
      }
    }

    if (url.pathname === "/refresh" && request.method === "POST") {
      if (!session?.tokens) return page("OIDC test failed", "<p>Complete login before rotating a refresh token.</p>");
      const previous = session.tokens;
      const { response, tokens, request: refreshRequest } = await tokenRequest({ grant_type: "refresh_token", client_id: clientId, refresh_token: previous.refresh_token });
      const [oldAccess, reusedRefresh] = await Promise.all([userInfo(previous.access_token), tokenRequest({ grant_type: "refresh_token", client_id: clientId, refresh_token: previous.refresh_token })]);
      if (!response.ok || !tokens.access_token || !tokens.refresh_token || !tokens.id_token || oldAccess.response.ok || reusedRefresh.response.ok) return page("Refresh rotation failed", `<pre>${escapeHtml(JSON.stringify({ refresh: tokens, oldAccessStatus: oldAccess.response.status, reusedRefresh: reusedRefresh.tokens }, null, 2))}</pre>`, session);
      const newAccess = await userInfo(tokens.access_token);
      if (!newAccess.response.ok) return page("Refresh rotation failed", "<p>The replacement access token was rejected by UserInfo.</p>", session);
      session.tokens = tokens as TokenSet;
      session.stage = 5;
      session.path = "rotated";
      session.events.push({ title: "Refresh Token rotation succeeded", details: { request: refreshRequest, status: response.status, tokenResponse: tokens } }, { title: "Replaced credentials were rejected", details: { previousAccessTokenRequest: oldAccess.request, previousAccessTokenUserInfoStatus: oldAccess.response.status, reusedRefreshTokenRequest: reusedRefresh.request, reusedRefreshTokenResponse: reusedRefresh.tokens } }, { title: "Replacement access token was accepted", details: { request: newAccess.request, userInfoStatus: newAccess.response.status } });
      return page("Refresh rotation passed", `<p>New access token: <code>${escapeHtml(shortToken(tokens.access_token))}</code></p>`, session);
    }

    if (url.pathname === "/revoke" && request.method === "POST") {
      if (!session?.tokens) return page("OIDC test failed", "<p>Complete login before revoking a token.</p>");
      const current = session.tokens;
      const revokeBody = new URLSearchParams({ client_id: clientId, token: current.refresh_token, token_type_hint: "refresh_token" });
      const revokeRequest = { method: "POST", url: `${issuerInternal}/revocation`, headers: { "content-type": "application/x-www-form-urlencoded" }, body: Object.fromEntries(revokeBody) };
      const revoke = await fetch(revokeRequest.url, { method: revokeRequest.method, headers: revokeRequest.headers, body: revokeBody });
      const [revokedAccess, revokedRefresh] = await Promise.all([userInfo(current.access_token), tokenRequest({ grant_type: "refresh_token", client_id: clientId, refresh_token: current.refresh_token })]);
      if (!revoke.ok || revokedAccess.response.ok || revokedRefresh.response.ok) return page("Revocation failed", `<pre>${escapeHtml(JSON.stringify({ revocationStatus: revoke.status, accessStatus: revokedAccess.response.status, refresh: revokedRefresh.tokens }, null, 2))}</pre>`, session);
      session.stage = 6;
      session.path ??= "direct";
      session.events.push({ title: "RFC 7009 revocation succeeded", details: { request: revokeRequest, status: revoke.status } }, { title: "Revoked tokens were rejected", details: { accessTokenRequest: revokedAccess.request, accessTokenUserInfoStatus: revokedAccess.response.status, refreshTokenRequest: revokedRefresh.request, refreshTokenResponse: revokedRefresh.tokens } });
      return page("Revocation passed", "<p>The current token set can no longer be used.</p>", session);
    }

    const initial = session ? `<p>Current flow state is retained in this browser session.</p>` : `<p>Issuer: ${escapeHtml(issuer)}</p><p>Client ID: ${escapeHtml(clientId)}</p><noscript><p><a href="/login">Start OIDC login without flow visualization</a></p></noscript>`;
    return page("UniQUE OIDC Test Client", initial, session);
  },
});
