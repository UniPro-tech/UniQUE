"use server";
import { forbidden, unauthorized } from "next/navigation";
import { Session } from "@/classes/Session";

/**
 * 権限がない場合にエラーをスロー
 * @param requiredFlag - 必要な権限フラグ
 * @param userRoles - ユーザーのロール一覧（省略時は自動取得）
 */
export async function requirePermission(requiredFlag: bigint): Promise<void> {
  const session = await Session.getCurrent();
  if (!session) {
    unauthorized();
  }
  const user = await session.getUser();
  if (!(await user.hasPermission(requiredFlag))) {
    forbidden();
  }
}
