"use server";
import { cookies } from "next/headers";
import { AuthenticationErrors } from "@/errors/AuthenticationErrors";
import { AuthenticationRequest, type Credentials } from "@/libs/authentication";
import { SetSessionCookie } from "@/libs/cookies";

export const submitSignIn = async (formData: FormData) => {
  return signIn(formData, "password");
};

export const submitMfa = async (formData: FormData) => {
  return signIn(formData, "totp");
};

const signIn = async (formData: FormData, type: Credentials["type"]) => {
  const username = formData.get("username") as string;
  const password = formData.get("password") as string | undefined;
  const is_remember = formData.get("remember") === "on";
  const code = formData.get("code") as string | undefined;
  const rawRedirectTo = formData.get("redirectTo") as string | undefined;
  const redirectTo =
    typeof rawRedirectTo === "string" &&
    rawRedirectTo.startsWith("/") &&
    !rawRedirectTo.startsWith("//")
      ? rawRedirectTo
      : "/dashboard";

  const credentials: Credentials = {
    type,
    username,
    password,
    code,
    is_remember,
  };

  try {
    const cookieStore = await cookies();
    if (type === "totp") {
      credentials.mfaChallenge = cookieStore.get(
        "pending_mfa_challenge",
      )?.value;
      credentials.username = cookieStore.get("pending_mfa_user")?.value;
      credentials.password = undefined;
      if (!credentials.mfaChallenge || !credentials.username) {
        throw AuthenticationErrors.InvalidCredentials;
      }
    } else {
      cookieStore.delete("pending_mfa_challenge");
      cookieStore.delete("pending_mfa_user");
      cookieStore.delete("pending_mfa_remember");
    }
    const response = await AuthenticationRequest(credentials);

    if (response.requireMfa) {
      if (type !== "password" || !response.mfaChallenge) {
        throw AuthenticationErrors.InvalidCredentials;
      }
      cookieStore.set("pending_mfa_challenge", response.mfaChallenge, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        path: "/",
        maxAge: 5 * 60,
      });
      cookieStore.set("pending_mfa_user", username, {
        path: "/",
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        maxAge: 5 * 60, // 5分間有効
      });

      cookieStore.set("pending_mfa_remember", is_remember ? "1" : "0", {
        path: "/",
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

    if (!response.sessionJwt) throw AuthenticationErrors.InvalidCredentials;
    await SetSessionCookie(response);
    cookieStore.delete("pending_mfa_challenge");
    cookieStore.delete("pending_mfa_user");
    cookieStore.delete("pending_mfa_remember");
    return { success: true, redirectTo: redirectTo || "/dashboard" };
  } catch (error) {
    const errorCodeMatch = (error as Error).message.match(/\[(.*?)\]/);
    const errorCode = errorCodeMatch ? errorCodeMatch[1] : "E0001";
    const queryParams = new URLSearchParams({
      username,
      remember: is_remember ? "1" : "0",
      error: errorCode,
      ...(redirectTo ? { redirect: redirectTo } : {}),
    });
    return { success: false, redirectTo: `/signin?${queryParams.toString()}` };
  }
};
