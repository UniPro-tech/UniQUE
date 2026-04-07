import { type NextRequest, NextResponse } from "next/server";
import { VerifyJwt } from "./libs/jwt";

const COOKIE_NAME = "session_jwt";

// This function can be marked `async` if using `await` inside
export async function proxy(request: NextRequest) {
  const jwtToken = request.cookies.get(COOKIE_NAME)?.value;
  if (!jwtToken) {
    // 旧Cookieが残っていたら削除
    const oldSid = request.cookies.get("unique-sid")?.value;
    const { pathname, search } = request.nextUrl;
    const redirectPath = `${pathname}${search}`;
    const redirectUrl = new URL("/signin", request.url);
    redirectUrl.searchParams.set("redirect", redirectPath);
    const response = NextResponse.redirect(redirectUrl);
    if (oldSid) {
      response.cookies.delete("unique-sid");
    }
    return response;
  }

  // JWT を検証
  const { valid, payload } = await VerifyJwt(jwtToken);
  if (!valid) {
    // JWT 検証失敗時は /signin にリダイレクト
    const { pathname, search } = request.nextUrl;
    const redirectPath = `${pathname}${search}`;
    const redirectUrl = new URL("/signin", request.url);
    redirectUrl.searchParams.set("redirect", redirectPath);
    const response = NextResponse.redirect(redirectUrl);
    response.cookies.delete(COOKIE_NAME);
    return response;
  }

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-url", request.url);
  // JWT のペイロード情報をヘッダーに追加（オプション）
  if (payload && typeof payload.sub === "string") {
    requestHeaders.set("x-user-id", payload.sub);
  }
  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  });
}

export const config = {
  matcher:
    "/((?!api|_next/static|_next/image|favicon.ico|signin|signup|migrate|terms|privacy|club_statute|email-verify|password-reset|.*\\.png|.*\\.webp|.*\\.svg).*)",
};
