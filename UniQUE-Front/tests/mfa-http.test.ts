import { expect, test } from "bun:test";

// Run against the production build and the optional Go test fixture server.
// Never print bodies, cookies, credentials, or the build's encryption key.
test.skipIf(!process.env.MFA_TEST_FRONT_URL || !process.env.MFA_TEST_AUTH_URL)(
  "HTTP Server Actions reject TOTP-only and complete password-bound MFA",
  async () => {
    const base = process.env.MFA_TEST_FRONT_URL as string;
    const auth = process.env.MFA_TEST_AUTH_URL as string;
    const manifest = await Bun.file(
      ".next/server/server-reference-manifest.json",
    ).json();
    const actionID = (name: string) => {
      const entry = Object.entries(manifest.node).find(
        ([, value]) =>
          (value as { exportedName?: string }).exportedName === name,
      );
      if (!entry) throw new Error("Build is missing the required action");
      return entry[0];
    };
    const fixture = await (await fetch(`${auth}/test/fixture`)).json();
    const count = async () =>
      (await (await fetch(`${auth}/test/sessions`)).json()).count;
    const send = async (
      action: string,
      fields: Record<string, string>,
      cookie = "",
    ) => {
      const body = new FormData();
      for (const [key, value] of Object.entries(fields))
        body.set(`_1_${key}`, value);
      body.set("0", '["$K1"]');
      return fetch(`${base}/signin/mfa?redirect=%2Fdashboard`, {
        method: "POST",
        body,
        redirect: "manual",
        headers: {
          "next-action": actionID(action),
          Accept: "text/x-component",
          Origin: base,
          ...(cookie ? { Cookie: cookie } : {}),
        },
      });
    };
    const before = await count();
    const unauthenticatedPage = await fetch(`${base}/device/consented`, {
      redirect: "manual",
    });
    expect(unauthenticatedPage.status).toBe(307);
    for (const action of ["submitMfa", "submitSignIn"]) {
      const rejected = await send(action, {
        username: fixture.username,
        code: fixture.code,
      });
      expect(
        rejected.headers
          .getSetCookie()
          .some((v) => v.startsWith("session_jwt=")),
      ).toBe(false);
      expect((await rejected.text()).includes('"success":true')).toBe(false);
      expect(await count()).toBe(before);
    }
    const password = await send("submitSignIn", {
      username: fixture.username,
      password: fixture.password,
      redirectTo: "/dashboard",
    });
    expect(password.status).toBe(200);
    const passwordBody = await password.text();
    const pending = password.headers.getSetCookie();
    expect(
      pending.some(
        (v) =>
          v.startsWith("pending_mfa_challenge=") &&
          !v.startsWith("pending_mfa_challenge=;"),
      ),
    ).toBe(true);
    expect(pending.some((v) => v.startsWith("session_jwt="))).toBe(false);
    expect(passwordBody.includes("/signin/mfa?redirect=%2Fdashboard")).toBe(
      true,
    );
    const cookie = pending.map((v) => v.split(";")[0]).join("; ");
    const completed = await send(
      "submitMfa",
      { code: fixture.code, redirectTo: "/dashboard" },
      cookie,
    );
    expect(
      completed.headers
        .getSetCookie()
        .some((v) => v.startsWith("session_jwt=") && v.includes("HttpOnly")),
    ).toBe(true);
    expect((await completed.text()).includes('"success":true')).toBe(true);
    expect(await count()).toBe(before + 1);
    const sessionCookie = completed.headers
      .getSetCookie()
      .find((v) => v.startsWith("session_jwt="))
      ?.split(";")[0] as string;
    const protectedPage = await fetch(`${base}/device/consented`, {
      headers: { Cookie: sessionCookie },
      redirect: "manual",
    });
    expect(protectedPage.status).toBe(200);
    const replay = await send("submitMfa", { code: fixture.code }, cookie);
    expect(
      replay.headers.getSetCookie().some((v) => v.startsWith("session_jwt=")),
    ).toBe(false);
    expect((await replay.text()).includes('"success":true')).toBe(false);
    expect(await count()).toBe(before + 1);
  },
);
