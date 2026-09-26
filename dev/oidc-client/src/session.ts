import type { LoginState, Prompt } from "./types";

const encoder = new TextEncoder();
const sessions = new Map<string, LoginState>();

const base64Url = (input: Uint8Array) =>
  Buffer.from(input).toString("base64url");

export const randomValue = () =>
  base64Url(crypto.getRandomValues(new Uint8Array(32)));

export const sha256 = async (value: string) =>
  new Uint8Array(
    await crypto.subtle.digest("SHA-256", encoder.encode(value)),
  );

export const readSession = (request: Request) => {
  const cookies = Object.fromEntries(
    (request.headers.get("cookie") ?? "")
      .split(";")
      .map((value) => value.trim().split("=", 2))
      .filter(([key]) => key),
  );
  const id = cookies.oidc_test_session;
  return { id, session: id ? sessions.get(id) : undefined };
};

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

export const deleteSession = (id?: string) => {
  if (id) sessions.delete(id);
};

export const sessionCookie = (id: string) =>
  `oidc_test_session=${id}; HttpOnly; SameSite=Lax; Path=/`;

export const expiredSessionCookie =
  "oidc_test_session=; Max-Age=0; HttpOnly; SameSite=Lax; Path=/";
