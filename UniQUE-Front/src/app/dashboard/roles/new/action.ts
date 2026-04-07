"use server";

import { redirect } from "next/navigation";
import { Role } from "@/classes/Role";

export interface CreateRoleFormData {
  customId: string;
  name: string;
  description: string;
  permissionBitmask: bigint;
  isDefault?: boolean;
  assignExisting?: boolean;
}

export async function createRole(
  data: CreateRoleFormData,
): Promise<{ success: boolean; error?: string; roleId?: string }> {
  try {
    const role = await Role.create({
      customId: data.customId,
      name: data.name,
      description: data.description || "",
      permissionBitmask: data.permissionBitmask || 0n,
      isDefault: data.isDefault || false,
    });

    if (data.assignExisting) {
      await role.asignAllUsers();
    }

    return { success: true, roleId: role.id };
  } catch (error) {
    console.error("Role creation failed:", error);
    return {
      success: false,
      error: "ロールの作成に失敗しました",
    };
  }
}

export async function redirectToRoleList() {
  redirect("/dashboard/roles");
}
