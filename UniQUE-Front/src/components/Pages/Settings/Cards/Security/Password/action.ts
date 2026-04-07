"use server";

import { Session } from "@/classes/Session";
import { AuthorizationErrors } from "@/errors/AuthorizationErrors";
import { FormRequestErrors } from "@/errors/FormRequestErrors";
import { VerifyCSRFToken } from "@/libs/csrf";
import type { FormStatus } from "../../Base";

export const updateSettings = async (
  _prevState: null | FormStatus,
  formData: FormData,
) => {
  try {
    // フォームデータから必要な情報を取得して処理を行う
    const csrfToken = formData.get("csrfToken") as string;
    const isVerified = await VerifyCSRFToken(csrfToken);
    if (!isVerified) throw FormRequestErrors.CSRFTokenMismatch;

    const currentPassword = formData.get("current_password") as string;
    const newPassword = formData.get("new_password") as string;
    const confirmNewPassword = formData.get("confirm_new_password") as string;

    if (newPassword !== confirmNewPassword) {
      return {
        status: "error",
        message: "新しいパスワードと確認欄が一致しません。",
      } as FormStatus;
    }

    const session = await Session.getCurrent();
    const user = await session?.getUser();
    if (!user) {
      return {
        status: "error",
        message: "ユーザーが見つかりませんでした。",
      } as FormStatus;
    }
    await user.changePassword(currentPassword, newPassword);

    // 必要に応じて結果を返す
    return {
      status: "success",
      message: "パスワードが更新されました。",
    } as FormStatus;
  } catch (error) {
    switch (error) {
      case FormRequestErrors.CSRFTokenMismatch:
        return {
          status: "error",
          message: "不正なリクエストです。ページをリロードしてください。",
        } as FormStatus;
      case AuthorizationErrors.Unauthorized:
        return {
          status: "error",
          message:
            "現在のパスワードが間違っている可能性があります。再度ご確認ください。",
        } as FormStatus;
      case AuthorizationErrors.AccessDenied:
        return {
          status: "error",
          message:
            "現在のパスワードが間違っている可能性があります。再度お試しください。",
        } as FormStatus;
      default:
        return {
          status: "error",
          message:
            "予期しないエラーが発生しました。時間をおいて再度お試しください。",
        } as FormStatus;
    }
  }
};
