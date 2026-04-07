"use server";

import { User } from "@/classes/User";

export async function submitForgotPassword(email: string) {
  if (typeof email !== "string") {
    return {
      status: "error",
      message: "Invalid email address.",
    } as const;
  }

  try {
    await User.passwordResetRequest(email);
    return {
      status: "success",
      message: "パスワードリセットメールを送信しました。",
    } as const;
  } catch (_error) {
    return {
      status: "error",
      message: "パスワードリセットに失敗しました。再度お試しください。",
    } as const;
  }
}
