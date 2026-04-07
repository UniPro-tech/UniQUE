"use server";
import { User } from "@/classes/User";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { FrontendErrors } from "@/errors/FrontendErrors";

export const userIdChangeApplyAction = async (
  _prevState: FormStatus | null,
  formData: FormData | null,
): Promise<FormStatus | null> => {
  try {
    if (!formData) {
      throw FrontendErrors.InvalidInput;
    }

    const userId = formData.get("userId");
    const newCustomId = formData.get("newCustomId");
    const changeReason = formData.get("changeReason");

    if (
      typeof userId !== "string" ||
      typeof newCustomId !== "string" ||
      typeof changeReason !== "string"
    ) {
      throw FrontendErrors.InvalidInput;
    }

    const user = await User.getById(userId);
    if (!user) {
      throw FrontendErrors.InvalidInput;
    }

    // TODO: 管理者への通知ロジックを追加する

    return {
      status: "success",
      message:
        "Not implemented: カスタムID変更申請が送信されました。管理者の承認をお待ちください。",
    };
  } catch (error) {
    switch (error) {
      case FrontendErrors.InvalidInput:
        return {
          status: "error",
          message: "無効な入力です。もう一度お試しください。",
        };
      default:
        console.error("Unexpected error in userIdChangeApplyAction:", error);
        return {
          status: "error",
          message: "予期しないエラーが発生しました。",
        };
    }
  }
};
