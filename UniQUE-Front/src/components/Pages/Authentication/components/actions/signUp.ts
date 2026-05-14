"use server";
import { User } from "@/classes/User";

export const submitSignUp = async (formData: FormData): Promise<string> => {
  console.log("Submitting sign-up form...");
  const external_email = formData.get("external_email") as string;
  const username = formData.get("username") as string;
  const name = formData.get("name") as string;
  const password = formData.get("password") as string;
  const passwordConfirm = formData.get("confirm_password") as string;
  const remember = formData.get("remember") === "on";

  if (password !== passwordConfirm) {
    const queryParams = new URLSearchParams({
      username,
      name,
      external_email,
      remember: remember ? "1" : "0",
    });
    return `/signup?${queryParams.toString()}`;
  }

  try {
    await User.create(
      {
        email: `temp_${Date.now()}@uniproject.jp`,
        externalEmail: external_email,
        customId: username,
        profile: {
          displayName: name,
        },
      },
      password,
    );
    return `/signup/success`;
  } catch (error) {
    const message = error instanceof Error ? error.message : "";
    const errorCodeMatch = message.match(/\[(.*?)\]/);
    const errorCode = errorCodeMatch ? errorCodeMatch[1] : "E0001";
    const queryParams = new URLSearchParams({
      username,
      name,
      external_email,
      remember: remember ? "1" : "0",
      error: errorCode,
    });
    return `/signup?${queryParams.toString()}`;
  }
};
