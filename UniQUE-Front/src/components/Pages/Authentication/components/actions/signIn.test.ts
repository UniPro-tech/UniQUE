import { beforeEach, expect, mock, test } from "bun:test";

const realFetch = globalThis.fetch;

const jar = new Map<string, string>();
const options = new Map<string, Record<string, unknown>>();
mock.module("next/headers", () => ({
  cookies: async () => ({
    get: (name: string) =>
      jar.has(name) ? { value: jar.get(name) } : undefined,
    getAll: () => [...jar].map(([name, value]) => ({ name, value })),
    set: (name: string, value: string, opts: Record<string, unknown>) => {
      jar.set(name, value);
      options.set(name, opts);
    },
    delete: (name: string) => {
      jar.delete(name);
    },
  }),
  headers: async () => new Headers(),
}));
// Exercise the real request serializer, action and session-cookie writer.
const { submitSignIn, submitMfa } = await import("./signIn");
let requests: Record<string, unknown>[] = [];
let response: Record<string, unknown>;
let status: number;
beforeEach(() => {
  jar.clear();
  options.clear();
  requests = [];
  status = 200;
  response = {};
  globalThis.fetch = mock(async (_url: unknown, init?: RequestInit) => {
    requests.push(JSON.parse(init?.body as string));
    return Response.json(response, { status });
  }) as unknown as typeof fetch;
});
function form(values: Record<string, string>) {
  const data = new FormData();
  for (const [key, value] of Object.entries(values)) data.set(key, value);
  return data;
}

test("T13 cookie-less direct MFA action cannot request or issue a session", async () => {
  const result = await submitMfa(form({ username: "test", code: "123456" }));
  expect(result.success).toBe(false);
  expect(requests).toHaveLength(0);
  expect(jar.has("session_jwt")).toBe(false);
});

test("old sign-in action never selects TOTP implicitly", async () => {
  status = 401;
  response = { reason: "invalid_credentials" };
  const result = await submitSignIn(form({ username: "test", code: "123456" }));
  expect(result.success).toBe(false);
  expect(requests[0].type).toBe("password");
  expect(jar.has("session_jwt")).toBe(false);
});

test("T14/T15 challenge cookie, bound username, session and OIDC redirect", async () => {
  response = { require_mfa: true, mfa_challenge: "local-test-challenge" };
  const redirect = "/authorization?request_id=local-test";
  const first = await submitSignIn(
    form({ username: "test", password: "password", redirectTo: redirect }),
  );
  expect(first.success).toBe(false);
  expect(first.redirectTo).toBe(
    `/signin/mfa?redirect=${encodeURIComponent(redirect)}`,
  );
  expect(jar.get("pending_mfa_challenge")).toBe("local-test-challenge");
  expect(options.get("pending_mfa_challenge")).toMatchObject({
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 300,
  });
  expect(jar.has("session_jwt")).toBe(false);
  // Synthetic unsigned fixture only; Auth's JWT issuance is covered in Go.
  const payload = btoa(
    JSON.stringify({ exp: Math.floor(Date.now() / 1000) + 3600 }),
  );
  response = { session_jwt: `test.${payload}.test` };
  const second = await submitMfa(
    form({ username: "other", code: "123456", redirectTo: redirect }),
  );
  expect(requests[1]).toMatchObject({
    type: "totp",
    username: "test",
    mfa_challenge: "local-test-challenge",
  });
  expect(second).toEqual({ success: true, redirectTo: redirect });
  expect(jar.has("session_jwt")).toBe(true);
  expect(jar.has("pending_mfa_challenge")).toBe(false);
  expect(jar.has("pending_mfa_user")).toBe(false);
});

test("Auth rejection with forged cookie does not set a session", async () => {
  jar.set("pending_mfa_challenge", "forged");
  jar.set("pending_mfa_user", "test");
  status = 401;
  response = { reason: "invalid_credentials" };
  expect((await submitMfa(form({ code: "123456" }))).success).toBe(false);
  expect(jar.has("session_jwt")).toBe(false);
});

test("incomplete Auth success fails closed", async () => {
  response = { require_mfa: true };
  expect(
    (await submitSignIn(form({ username: "test", password: "password" })))
      .success,
  ).toBe(false);
  expect(jar.has("pending_mfa_challenge")).toBe(false);
  response = {};
  expect(
    (await submitSignIn(form({ username: "test", password: "password" })))
      .success,
  ).toBe(false);
  expect(jar.has("session_jwt")).toBe(false);
});

test.skipIf(!process.env.MFA_TEST_AUTH_URL)(
  "live Auth: TOTP-only denied; password + challenge succeeds once",
  async () => {
    const base = process.env.MFA_TEST_AUTH_URL as string;
    process.env.AUTH_API_URL = base;
    globalThis.fetch = realFetch;
    const fixture = (await (
      await realFetch(`${base}/test/fixture`)
    ).json()) as Record<string, string>;
    const count = async () =>
      (
        (await (await realFetch(`${base}/test/sessions`)).json()) as {
          count: number;
        }
      ).count;
    const before = await count();
    // A currently valid TOTP, no password and no cookies, through both Actions.
    expect(
      (
        await submitMfa(
          form({ username: fixture.username, code: fixture.code }),
        )
      ).success,
    ).toBe(false);
    expect(
      (
        await submitSignIn(
          form({ username: fixture.username, code: fixture.code }),
        )
      ).success,
    ).toBe(false);
    expect(jar.has("session_jwt")).toBe(false);
    expect(await count()).toBe(before);
    // Also exercise the Auth boundary directly, without any Front validation.
    const bypass = await realFetch(`${base}/internal/authentication`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        type: "totp",
        username: fixture.username,
        code: fixture.code,
      }),
    });
    expect(bypass.status).toBe(401);
    expect(
      ((await bypass.json()) as { session_jwt?: string }).session_jwt,
    ).toBeUndefined();
    expect(await count()).toBe(before);
    const first = await submitSignIn(
      form({ username: fixture.username, password: fixture.password }),
    );
    expect(first.redirectTo).toBe("/signin/mfa?redirect=%2Fdashboard");
    const challenge = jar.get("pending_mfa_challenge") as string;
    expect(challenge.length).toBe(43);
    expect(jar.has("session_jwt")).toBe(false);
    expect((await submitMfa(form({ code: fixture.code }))).success).toBe(true);
    expect(jar.has("session_jwt")).toBe(true);
    expect(options.get("session_jwt")?.httpOnly).toBe(true);
    expect(await count()).toBe(before + 1);
    const jwt = jar.get("session_jwt") as string;
    const claims = JSON.parse(atob(jwt.split(".")[1]));
    const verification = await realFetch(
      `${base}/internal/session_verify?jti=${encodeURIComponent(claims.sub.slice(4))}`,
    );
    expect(((await verification.json()) as { valid: boolean }).valid).toBe(
      true,
    );
    jar.delete("session_jwt");
    jar.set("pending_mfa_challenge", challenge);
    jar.set("pending_mfa_user", fixture.username);
    expect((await submitMfa(form({ code: fixture.code }))).success).toBe(false);
    expect(jar.has("session_jwt")).toBe(false);
    expect(await count()).toBe(before + 1);
  },
);
