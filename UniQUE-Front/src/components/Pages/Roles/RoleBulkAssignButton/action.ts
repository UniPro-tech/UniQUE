"use server";

import { Role } from "@/classes/Role";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export const bulkAssignUsersToRole = async ({
  roleId,
}: {
  roleId: string;
}): Promise<{ success: boolean; error?: string }> => {
  const role = await Role.getById(roleId);
  if (!role) {
    throw ResourceApiErrors.ResourceNotFound;
  }
  await role?.asignAllUsers();
  return { success: true };
};
