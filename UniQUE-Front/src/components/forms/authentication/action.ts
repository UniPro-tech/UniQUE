"use server";

import type { UserData } from "@/classes/types/User";
import { User } from "@/classes/User";
import type { UniQUE_Error } from "@/errors/base";

export interface SignupFormState {
  customId: string;
  externalEmail: string;
  displayName: string;
  password: string;
  confirmPassword: string;
  birthdate: string;
}

export const signupAction = async (
  data: SignupFormState,
): Promise<Error | UniQUE_Error | UserData> => {
  const res = await User.create(
    {
      customId: data.customId,
      email: `tmp_${new Date().getUTCMilliseconds()}@uniproject.jp`,
      externalEmail: data.externalEmail,
      profile: {
        displayName: data.displayName,
        birthdate: data.birthdate,
      },
    },
    data.password,
  );
  if (res instanceof User) {
    return res.toJson();
  }
  return res;
};
