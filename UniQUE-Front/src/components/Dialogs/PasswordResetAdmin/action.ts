"use server";

import { User } from "@/classes/User";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export const resetPassword = async (userId: string, newPassword: string) => {
  const user = await User.getById(userId);
  if (!user) throw ResourceApiErrors.ResourceNotFound;
  await user.changePassword(null, newPassword);
};
