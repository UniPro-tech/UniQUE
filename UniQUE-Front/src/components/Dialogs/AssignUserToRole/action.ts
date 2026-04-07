"use server";

import { Role } from "@/classes/Role";
import type { UserData } from "@/classes/types/User";
import { User } from "@/classes/User";

export async function getAllUsers(): Promise<UserData[]> {
  try {
    const users = await User.getAll();
    return users.map((u) => u.toJson());
  } catch (error) {
    console.error("Failed to fetch users:", error);
    return [];
  }
}

export async function assignUserToRole(
  roleId: string,
  userId: string,
): Promise<{ success: boolean; error?: string }> {
  try {
    const role = await Role.getById(roleId);
    await role?.addUser(userId);
    return { success: true };
  } catch (error) {
    console.error("Failed to assign user to role:", error);
    return {
      success: false,
      error: "ユーザーの割り当てに失敗しました",
    };
  }
}

export async function unassignUserFromRole(
  roleId: string,
  userId: string,
): Promise<{ success: boolean; error?: string }> {
  try {
    const role = await Role.getById(roleId);
    await role?.removeUser(userId);
    return { success: true };
  } catch (error) {
    console.error("Failed to unassign user from role:", error);
    return {
      success: false,
      error: "ユーザーの割り当て解除に失敗しました",
    };
  }
}
