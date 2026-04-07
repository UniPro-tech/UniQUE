"use server";

import { User } from "@/classes/User";

export const passwordResetRequest = async (email: string) => {
  await User.passwordResetRequest(email);
};

export const confirmPasswordReset = async (code: string, password: string) => {
  return await User.confirmPasswordReset(code, password);
};
