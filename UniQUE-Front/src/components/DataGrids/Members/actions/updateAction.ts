"use server";

import type { UserData } from "@/classes/types/User";
import { User } from "@/classes/User";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export async function updateUserById(
  userId: string,
  data: Partial<Omit<UserData, "id" | "createdAt" | "updatedAt" | "deletedAt">>,
): Promise<UserData> {
  const user = await User.getById(userId);
  if (!user) {
    throw ResourceApiErrors.ResourceNotFound;
  }

  return (await User.updateById(userId, data)).toJson();
}
