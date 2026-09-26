import { clientId, issuer, issuerInternal, redirectUri } from "./config";
import { sha256 } from "./session";
import type { LoginState, RequestLog, TokenSet } from "./types";

const encoder = new TextEncoder();

const decodeJwtPart = (value: string) =>
  JSON.parse(Buffer.from(value, "base64url").toString());

export const authorizationRequest = async (session: LoginState) => {
  const parameters = new URLSearchParams({
    client_id: clientId,
    redirect_uri: redirectUri,
    response_type: "code",
    scope: "openid profile email",
    state: session.state,
    nonce: session.nonce,
    code_challenge: Buffer.from(await sha256(session.verifier)).toString(
      "base64url",
    ),
    code_challenge_method: "S256",
  });
  if (session.prompt) parameters.set("prompt", session.prompt);

  return {
    url: `${issuer}/authorization?${parameters}`,
    request: {
      method: "GET",
      url: `${issuer}/authorization?${parameters}`,
      body: Object.fromEntries(parameters),
    } satisfies RequestLog,
  };
};

export const verifyIdToken = async (token: string, nonce: string) => {
  const [encodedHeader, encodedClaims, encodedSignature] = token.split(".");
  if (!encodedHeader || !encodedClaims || !encodedSignature) {
    throw new Error("ID token is malformed");
  }

  const header = decodeJwtPart(encodedHeader) as { kid?: string; alg?: string };
  const claims = decodeJwtPart(encodedClaims) as {
    iss?: string;
    aud?: string | string[];
    nonce?: string;
    exp?: number;
  };
  const jwks = (await (
    await fetch(`${issuerInternal}/.well-known/jwks.json`)
  ).json()) as { keys: JsonWebKey[] };
  const key = jwks.keys.find((candidate) => candidate.kid === header.kid);
  if (!key || header.alg !== "RS256") {
    throw new Error("matching RS256 JWKS key was not found");
  }

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
  if (
    !validSignature ||
    claims.iss !== issuer ||
    !audience.includes(clientId) ||
    claims.nonce !== nonce ||
    !claims.exp ||
    claims.exp * 1000 < Date.now()
  ) {
    throw new Error("ID token validation failed");
  }
  return claims;
};

export const tokenRequest = async (parameters: Record<string, string>) => {
  const body = new URLSearchParams(parameters);
  const request: RequestLog = {
    method: "POST",
    url: `${issuerInternal}/token`,
    headers: { "content-type": "application/x-www-form-urlencoded" },
    body: Object.fromEntries(body),
  };
  const response = await fetch(request.url, {
    method: request.method,
    headers: request.headers,
    body,
  });
  const tokens = (await response.json()) as Partial<TokenSet> & {
    error?: string;
  };
  return { response, tokens, request };
};

export const userInfo = async (accessToken: string) => {
  const request: RequestLog = {
    method: "GET",
    url: `${issuerInternal}/userinfo`,
    headers: { authorization: `Bearer ${accessToken}` },
  };
  const response = await fetch(request.url, { headers: request.headers });
  return { response, body: await response.json(), request };
};

export const revokeToken = async (refreshToken: string) => {
  const body = new URLSearchParams({
    client_id: clientId,
    token: refreshToken,
    token_type_hint: "refresh_token",
  });
  const request: RequestLog = {
    method: "POST",
    url: `${issuerInternal}/revocation`,
    headers: { "content-type": "application/x-www-form-urlencoded" },
    body: Object.fromEntries(body),
  };
  const response = await fetch(request.url, {
    method: request.method,
    headers: request.headers,
    body,
  });
  return { response, request };
};

export const codeExchangeParameters = (code: string, verifier: string) => ({
  grant_type: "authorization_code",
  client_id: clientId,
  code,
  redirect_uri: redirectUri,
  code_verifier: verifier,
});

export const refreshParameters = (refreshToken: string) => ({
  grant_type: "refresh_token",
  client_id: clientId,
  refresh_token: refreshToken,
});
