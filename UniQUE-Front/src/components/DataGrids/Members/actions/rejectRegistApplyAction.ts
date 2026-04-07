"use server";
import { User } from "@/classes/User";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";

export const rejectRegistApplyAction = async (
  _prevState: FormStatus | null,
  formData: FormData | null,
): Promise<FormStatus | null> => {
  if (!formData) {
    return { status: "error", message: "不正なフォームデータです。" };
  }
  const userId = formData.get("userId");
  if (!userId || typeof userId !== "string") {
    return { status: "error", message: "userId が指定されていません" };
  }

  try {
    const user = await User.getById(userId);
    if (!user) {
      return { status: "error", message: "ユーザーが見つかりませんでした" };
    }
    await user.reject();
    return { status: "success", message: "申請を却下しました" };
  } catch (err) {
    console.error("rejectRegistApplyAction error:", err);
    return { status: "error", message: "却下中にエラーが発生しました" };
  }
};
