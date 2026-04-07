"use server";
import { cookies } from "next/headers";
import { AuthenticationRequest, type Credentials } from "@/libs/authentication";
import { SetSessionCookie } from "@/libs/cookies";

export const submitSignIn = async (formData: FormData) => {
  const username = formData.get("username") as string;
  const password = formData.get("password") as string | undefined;
  const remember = formData.get("remember") === "on";
  const code = formData.get("code") as string | undefined;
  const rawRedirectTo = formData.get("redirectTo") as string | undefined;
  const redirectTo =
    typeof rawRedirectTo === "string" &&
    rawRedirectTo.startsWith("/") &&
    !rawRedirectTo.startsWith("//")
      ? rawRedirectTo
      : "/dashboard";

  const credentials: Credentials = {
    username,
    password,
    code,
    remember,
  };

  try {
    const response = await AuthenticationRequest(credentials);

    if (response.requireMfa) {
      const cookieStore = await cookies();
      cookieStore.set("pending_mfa_user", username, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 5 * 60, // 5分間有効
      });

      cookieStore.set("pending_mfa_remember", remember ? "1" : "0", {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 5 * 60, // 5分間有効
      });

      return {
        success: false,
        redirectTo: `/signin/mfa?redirect=${encodeURIComponent(redirectTo || "/dashboard")}`,
      };
    }

    await SetSessionCookie(response);
    return { success: true, redirectTo: redirectTo || "/dashboard" };
  } catch (error) {
    const errorCodeMatch = (error as Error).message.match(/\[(.*?)\]/);
    const errorCode = errorCodeMatch ? errorCodeMatch[1] : "E0001";
    const queryParams = new URLSearchParams({
      username,
      remember: remember ? "1" : "0",
      error: errorCode,
      ...(redirectTo ? { redirect: redirectTo } : {}),
    });
    return { success: false, redirectTo: `/signin?${queryParams.toString()}` };
  }
};
