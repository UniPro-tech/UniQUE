"use server";

import { Role } from "@/classes/Role";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export interface UpdateRoleFormData {
  customId: string;
  name: string;
  description: string;
  permissionBitmask: bigint;
  isDefault?: boolean;
}

export async function updateRole(
  roleId: string,
  data: UpdateRoleFormData,
): Promise<{ success: boolean; error?: string }> {
  try {
    const role = await Role.getById(roleId);
    if (!role) {
      throw ResourceApiErrors.ResourceNotFound;
    }

    role.customId = data.customId;
    role.name = data.name;
    role.description = data.description;
    role.permissionBitmask = data.permissionBitmask;
    if (typeof data.isDefault !== "undefined") {
      role.isDefault = data.isDefault;
    }

    await role.save();

    return { success: true };
  } catch (error) {
    console.error("Role update failed:", error);
    return {
      success: false,
      error: "ロールの更新に失敗しました",
    };
  }
}

export async function deleteRole(
  roleId: string,
): Promise<{ success: boolean; error?: string }> {
  try {
    await Role.deleteById(roleId);
    return { success: true };
  } catch (error) {
    console.error("Role delete failed:", error);
    return { success: false, error: "ロールの削除に失敗しました" };
  }
}
