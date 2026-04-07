"use server";

import { User } from "@/classes/User";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";

export const approveAction = async (
  _prevState: FormStatus | null,
  formData: FormData | null,
): Promise<FormStatus | null> => {
  const userId = formData?.get("userId");
  if (typeof userId !== "string") {
    return {
      status: "error",
      message: "カスタムIDが無効です。",
    };
  }
  const period = formData?.get("period");
  if (typeof period !== "string") {
    return {
      status: "error",
      message: "所属期が無効です。",
    };
  }
  const mailboxPassword = formData?.get("mailboxPassword");
  if (typeof mailboxPassword !== "string") {
    return {
      status: "error",
      message: "メールボックスのパスワードが無効です。",
    };
  }
  const email = formData?.get("email");
  if (typeof email !== "string") {
    return {
      status: "error",
      message: "メールアドレスが無効です。",
    };
  }
  try {
    const user = await User.getById(userId);
    if (!user) {
      return {
        status: "error",
        message: "ユーザーが見つかりませんでした。",
      };
    }
    await user.approve({
      affiliationPeriod: period,
      sakuraEmailPassword: mailboxPassword,
      email: email,
    });
    return {
      status: "success",
      message: "メンバーを承認しました。",
    };
  } catch (e) {
    console.log(e);
    return {
      status: "error",
      message: "メンバーの承認に失敗しました。",
    };
  }
  return null;
};
