"use server";

import { Session } from "@/classes/Session";

export const generateTOTP = async (_prevState: null, formData: FormData) => {
  try {
    const password = formData.get("password") as string;

    const session = await Session.getCurrent();
    const user = await session?.getUser();
    if (!user) {
      return { error: "ユーザーが見つかりませんでした" };
    }

    const totpRes = await user.setupBeginTotp(password);

    return totpRes;
  } catch (error) {
    return { error: error instanceof Error ? error.message : String(error) };
  }
};

export const verifyTOTP = async (_prevState: null, formData: FormData) => {
  try {
    const code = formData.get("code") as string;

    const session = await Session.getCurrent();

    const user = await session?.getUser();
    if (!user) {
      return { error: "ユーザーが見つかりませんでした" };
    }

    const result = await user.setupFinishTotp(code);

    return result;
  } catch (error) {
    return { error: error instanceof Error ? error.message : String(error) };
  }
};

export const disableTOTP = async (_prevState: null, formData: FormData) => {
  try {
    const password = formData.get("password") as string;

    const session = await Session.getCurrent();
    const user = await session?.getUser();
    if (!user) {
      return { error: "ユーザーが見つかりませんでした" };
    }

    await user.disableTotp(password);

    return { success: true, message: "TOTPが無効化されました。" };
  } catch (error) {
    return { error: error instanceof Error ? error.message : String(error) };
  }
};
