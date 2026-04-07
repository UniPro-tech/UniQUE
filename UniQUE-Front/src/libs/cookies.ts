import { cookies } from "next/headers";
import type { AuthenticationResponse } from "./authentication";

// サーバー/クライアント両対応の cookie 取得
export const getAllCookies = async (): Promise<string> => {
  // クライアントバンドルの場合は document.cookie を優先
  if (typeof window !== "undefined") {
    return typeof document !== "undefined" ? (document.cookie ?? "") : "";
  }

  // サーバーの場合のみ next/headers を動的インポート
  const { cookies } = await import("next/headers");
  const cookieStore = await cookies();
  const cookie = cookieStore
    .getAll()
    .map((c) => `${c.name}=${c.value}`)
    .join(";");
  return cookie;
};

export const ClearSessionCookie = async () => {
  const cookieStore = await cookies();
  cookieStore.delete("session_jwt");
};

export const SetSessionCookie = async (response: AuthenticationResponse) => {
  const { sessionJwt } = response;
  const cookieName = "session_jwt";
  const expires = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000); // セッション終了時

  const cookieStore = await cookies();
  cookieStore.set(cookieName, sessionJwt || "", {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    expires,
  });
};
