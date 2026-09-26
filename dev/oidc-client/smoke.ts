const issuer = process.env.OIDC_ISSUER ?? "http://localhost:8000";
const client = process.env.OIDC_CLIENT_URL ?? "http://localhost:3002";

function requireRedirect(response: Response) {
  const location = response.headers.get("location");
  if (response.status < 300 || response.status > 399 || !location) {
    throw new Error(`expected redirect, received ${response.status}`);
  }
  return location;
}

const login = await fetch(`${client}/login`, { redirect: "manual" });
const authorizationUrl = requireRedirect(login);
const sessionCookie = login.headers.get("set-cookie")?.match(/oidc_test_session=([^;]+)/)?.[1];
if (!sessionCookie) throw new Error("OIDC client did not issue a session cookie");

const authorization = await fetch(authorizationUrl, { redirect: "manual" });
const frontendUrl = requireRedirect(authorization);
const authorizationRequestID = new URL(frontendUrl).searchParams.get("auth_request_id");
if (!authorizationRequestID) throw new Error("authorization request ID was not returned");

const sessionResponse = await fetch(`${issuer}/internal/authentication`, {
  method: "POST",
  headers: { "content-type": "application/json" },
  body: JSON.stringify({ username: "test", password: "testpassword", type: "password" }),
});
const session = await sessionResponse.json() as { session_jwt?: string };
if (!sessionResponse.ok || !session.session_jwt) throw new Error("test-user authentication failed");

const consent = await fetch(`${issuer}/authorization`, {
  method: "POST",
  headers: { "content-type": "application/x-www-form-urlencoded" },
  body: new URLSearchParams({ auth_request_id: authorizationRequestID, session_jwt: session.session_jwt }),
  redirect: "manual",
});
const consentedUrl = requireRedirect(consent);

const codeRedirect = await fetch(consentedUrl, { redirect: "manual" });
const callbackUrl = requireRedirect(codeRedirect);
if (!callbackUrl.startsWith(`${client}/callback`)) throw new Error("authorization code was not redirected to the OIDC client");

const callback = await fetch(callbackUrl, {
  headers: { cookie: `oidc_test_session=${sessionCookie}` },
});
const body = await callback.text();
if (!callback.ok || !body.includes("OIDC test succeeded")) throw new Error(`OIDC callback failed: ${body}`);

const sessionHeaders = { cookie: `oidc_test_session=${sessionCookie}` };
const refresh = await fetch(`${client}/refresh`, { method: "POST", headers: sessionHeaders });
const refreshBody = await refresh.text();
if (!refresh.ok || !refreshBody.includes("Refresh rotation passed")) throw new Error(`refresh rotation failed: ${refreshBody}`);

const revocation = await fetch(`${client}/revoke`, { method: "POST", headers: sessionHeaders });
const revocationBody = await revocation.text();
if (!revocation.ok || !revocationBody.includes("Revocation passed")) throw new Error(`revocation failed: ${revocationBody}`);

console.log("OIDC authorization, refresh rotation, and revocation passed");
