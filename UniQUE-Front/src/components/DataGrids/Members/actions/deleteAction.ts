"use server";

import { User } from "@/classes/User";
import { logger } from "@/libs/logger";

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
    logger.error("Failed to delete user:", err);
    return { success: false, error: "ユーザーの削除に失敗しました" };
  }
}
