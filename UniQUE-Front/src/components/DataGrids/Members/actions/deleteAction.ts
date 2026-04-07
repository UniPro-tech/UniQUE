"use server";

import { User } from "@/classes/User";

export async function deleteUserById(
  userId: string,
): Promise<{ success: boolean; error?: string }> {
  try {
    const user = await User.getById(userId);
    if (!user) {
      return { success: false, error: "ユーザーが見つかりませんでした" };
    }
    await user.delete();
    return { success: true };
  } catch (err) {
    console.error("Failed to delete user:", err);
    return { success: false, error: "ユーザーの削除に失敗しました" };
  }
}
