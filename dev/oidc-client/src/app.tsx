/** @jsxImportSource hono/jsx */

import { Hono } from "hono";
import { serveStatic } from "hono/bun";
import {
  authorizationRequest,
  codeExchangeParameters,
  refreshParameters,
  revokeToken,
  tokenRequest,
  userInfo,
  verifyIdToken,
} from "./oidc";
import {
  createSession,
  deleteSession,
  expiredSessionCookie,
  readSession,
  sessionCookie,
} from "./session";
import type { LoginState, Prompt, TokenSet } from "./types";
import { ErrorPage, Home, Page, shortToken } from "./ui";

const app = new Hono();

const isPrompt = (value: string | undefined): value is Prompt =>
  value === "consent" || value === "none";

const hasValidState = (
  session: LoginState | undefined,
  state: string | undefined,
) => session && session.expiresAt >= Date.now() && state === session.state;

const renderError = (
  message: string,
  session?: LoginState,
  callbackError = false,
) => <ErrorPage message={message} session={session} callbackError={callbackError} />;

app.use("/styles.css", serveStatic({ path: "./public/styles.css" }));
app.get("/health", (c) => c.json({ status: "ok" }));

app.get("/login", async (c) => {
  const requestedPrompt = c.req.query("prompt");
  if (requestedPrompt && !isPrompt(requestedPrompt)) {
    return c.html(renderError("Unsupported prompt value."));
  }

  const { id: existingId, session: existingSession } = readSession(c.req.raw);
  const prompt = requestedPrompt ?? undefined;
  const reusable =
    existingId &&
    existingSession &&
    existingSession.expiresAt >= Date.now() &&
    existingSession.stage < 4 &&
    !existingSession.error &&
    existingSession.prompt === prompt;
  const active = reusable
    ? { id: existingId, session: existingSession }
    : createSession(prompt);
  const authorization = await authorizationRequest(active.session);

  if (!reusable) {
    active.session.events.push({
      title: "Authorization request prepared",
      details: { request: authorization.request },
    });
    c.header("Set-Cookie", sessionCookie(active.id));
  }
  return c.redirect(authorization.url, 302);
});

app.get("/callback", async (c) => {
  const { id, session } = readSession(c.req.raw);
  if (!id || !hasValidState(session, c.req.query("state"))) {
    return c.html(renderError("Missing, expired, or invalid state."));
  }
  if (session.tokens && !session.error) {
    return c.redirect("/", 303);
  }

  const code = c.req.query("code");
  if (!code) {
    const error = c.req.query("error") ?? "No authorization code returned";
    session.error = error;
    session.events.push({
      title: "Authorization response returned an error",
      details: { parameters: Object.fromEntries(new URL(c.req.url).searchParams) },
    });
    return c.html(renderError(error, session, true));
  }

  const token = await tokenRequest(codeExchangeParameters(code, session.verifier));
  if (
    !token.response.ok ||
    !token.tokens.id_token ||
    !token.tokens.access_token ||
    !token.tokens.refresh_token
  ) {
    session.error = "Token exchange failed";
    session.events.push({
      title: "Token exchange failed",
      details: { request: token.request, status: token.response.status, response: token.tokens },
    });
    return c.html(renderError(JSON.stringify(token.tokens, null, 2), session, true));
  }

  try {
    const [idToken, info] = await Promise.all([
      verifyIdToken(token.tokens.id_token, session.nonce),
      userInfo(token.tokens.access_token),
    ]);
    if (!info.response.ok) throw new Error("UserInfo rejected the new access token");

    session.tokens = token.tokens as TokenSet;
    session.stage = 4;
    session.events.push(
      {
        title: "Authorization code exchanged and ID Token verified",
        details: { request: token.request, status: token.response.status, idTokenClaims: idToken, tokenResponse: token.tokens },
      },
      {
        title: "UserInfo accepted the access token",
        details: { request: info.request, status: info.response.status, response: info.body },
      },
    );
    return c.html(
      <Page title="OIDC test succeeded" session={session} callbackComplete>
        <p>Authorization Code Flow with PKCE completed.</p>
        <h2>ID Token</h2>
        <pre>{JSON.stringify(idToken, null, 2)}</pre>
        <h2>UserInfo</h2>
        <pre>{JSON.stringify(info.body, null, 2)}</pre>
      </Page>,
    );
  } catch (error) {
    session.error = error instanceof Error ? error.message : String(error);
    session.events.push({ title: "Token validation failed", details: { error: session.error } });
    return c.html(renderError(session.error, session, true));
  }
});

app.post("/refresh", async (c) => {
  const { session } = readSession(c.req.raw);
  if (!session?.tokens) {
    return c.html(renderError("Complete login before rotating a refresh token."));
  }

  const previous = session.tokens;
  const refreshed = await tokenRequest(refreshParameters(previous.refresh_token));
  const [oldAccess, reusedRefresh] = await Promise.all([
    userInfo(previous.access_token),
    tokenRequest(refreshParameters(previous.refresh_token)),
  ]);
  if (
    !refreshed.response.ok ||
    !refreshed.tokens.access_token ||
    !refreshed.tokens.refresh_token ||
    !refreshed.tokens.id_token ||
    oldAccess.response.ok ||
    reusedRefresh.response.ok
  ) {
    if (
      refreshed.response.ok &&
      refreshed.tokens.access_token &&
      refreshed.tokens.refresh_token &&
      refreshed.tokens.id_token
    ) {
      session.tokens = refreshed.tokens as TokenSet;
    }
    return c.html(
      <Page title="Refresh rotation failed" session={session}>
        <pre>{JSON.stringify({ refresh: refreshed.tokens, oldAccessStatus: oldAccess.response.status, reusedRefresh: reusedRefresh.tokens }, null, 2)}</pre>
      </Page>,
    );
  }

  const newAccess = await userInfo(refreshed.tokens.access_token);
  if (!newAccess.response.ok) {
    return c.html(<Page title="Refresh rotation failed" session={session}><p>The replacement access token was rejected by UserInfo.</p></Page>);
  }

  session.tokens = refreshed.tokens as TokenSet;
  session.stage = 5;
  session.path = "rotated";
  session.events.push(
    {
      title: "Refresh Token rotation succeeded",
      details: { request: refreshed.request, status: refreshed.response.status, tokenResponse: refreshed.tokens },
    },
    {
      title: "Replaced credentials were rejected",
      details: {
        previousAccessTokenRequest: oldAccess.request,
        previousAccessTokenUserInfoStatus: oldAccess.response.status,
        reusedRefreshTokenRequest: reusedRefresh.request,
        reusedRefreshTokenResponse: reusedRefresh.tokens,
      },
    },
    {
      title: "Replacement access token was accepted",
      details: { request: newAccess.request, userInfoStatus: newAccess.response.status },
    },
  );
  return c.html(
    <Page title="Refresh rotation passed" session={session}>
      <p>New access token: <code>{shortToken(refreshed.tokens.access_token)}</code></p>
    </Page>,
  );
});

app.post("/revoke", async (c) => {
  const { session } = readSession(c.req.raw);
  if (!session?.tokens) {
    return c.html(renderError("Complete login before revoking a token."));
  }

  const current = session.tokens;
  const revocation = await revokeToken(current.refresh_token);
  if (!revocation.response.ok) {
    return c.html(
      <Page title="Revocation failed" session={session}>
        <pre>{JSON.stringify({ revocationStatus: revocation.response.status }, null, 2)}</pre>
      </Page>,
    );
  }
  const [revokedAccess, revokedRefresh] = await Promise.all([
    userInfo(current.access_token),
    tokenRequest(refreshParameters(current.refresh_token)),
  ]);
  if (revokedAccess.response.ok || revokedRefresh.response.ok) {
    return c.html(
      <Page title="Revocation failed" session={session}>
        <pre>{JSON.stringify({ revocationStatus: revocation.response.status, accessStatus: revokedAccess.response.status, refresh: revokedRefresh.tokens }, null, 2)}</pre>
      </Page>,
    );
  }

  session.stage = 6;
  session.path ??= "direct";
  session.events.push(
    {
      title: "RFC 7009 revocation succeeded",
      details: { request: revocation.request, status: revocation.response.status },
    },
    {
      title: "Revoked tokens were rejected",
      details: {
        accessTokenRequest: revokedAccess.request,
        accessTokenUserInfoStatus: revokedAccess.response.status,
        refreshTokenRequest: revokedRefresh.request,
        refreshTokenResponse: revokedRefresh.tokens,
      },
    },
  );
  return c.html(<Page title="Revocation passed" session={session}><p>The current token set can no longer be used.</p></Page>);
});

app.post("/reset", (c) => {
  const { id } = readSession(c.req.raw);
  deleteSession(id);
  c.header("Set-Cookie", expiredSessionCookie);
  return c.redirect("/", 303);
});

app.get("/", (c) => c.html(<Home session={readSession(c.req.raw).session} />));

export { app };
