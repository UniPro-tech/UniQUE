import type { LoginState, Prompt } from "./types";

const encoder = new TextEncoder();
const sessions = new Map<string, LoginState>();

/** Encodes bytes with unpadded Base64URL. */
const base64Url = (input: Uint8Array) =>
  Buffer.from(input).toString("base64url");

/** Generates a cryptographically random OIDC parameter. */
export const randomValue = () =>
  base64Url(crypto.getRandomValues(new Uint8Array(32)));

/** Computes the SHA-256 digest used by PKCE. */
export const sha256 = async (value: string) =>
  new Uint8Array(
    await crypto.subtle.digest("SHA-256", encoder.encode(value)),
  );

/** Reads the in-memory login session selected by the request cookie. */
export const readSession = (request: Request) => {
  const cookies = Object.fromEntries(
    (request.headers.get("cookie") ?? "")
      .split(";")
      .map((value) => value.trim().split("=", 2))
      .filter(([key]) => key),
  );
  const id = cookies.oidc_test_session;
  const session = id ? sessions.get(id) : undefined;
  if (id && session && session.expiresAt < Date.now()) {
    sessions.delete(id);
    return { id, session: undefined };
  }
  return { id, session };
};

/** Creates and stores a short-lived login session. */
export const createSession = (prompt?: Prompt) => {
  const id = randomValue();
  const session: LoginState = {
    state: randomValue(),
    nonce: randomValue(),
    verifier: randomValue(),
    expiresAt: Date.now() + 10 * 60 * 1000,
    stage: 1,
    prompt,
    events: [],
  };
  sessions.set(id, session);
  return { id, session };
};

/** Deletes an in-memory login session when an ID is present. */
export const deleteSession = (id?: string) => {
  if (id) sessions.delete(id);
};

/** Serializes the secure session identifier cookie. */
export const sessionCookie = (id: string) =>
  `oidc_test_session=${id}; HttpOnly; SameSite=Lax; Path=/`;

export const expiredSessionCookie =
  "oidc_test_session=; Max-Age=0; HttpOnly; SameSite=Lax; Path=/";
