"use server";
import { RedirectType, redirect } from "next/navigation";
import { User } from "@/classes/User";

export const submitSignUp = async (formData: FormData) => {
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
    redirect(`/signup?${queryParams.toString()}`, RedirectType.replace);
  }

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
  ).catch((error) => {
    const errorCodeMatch = error.message.match(/\[(.*?)\]/);
    const errorCode = errorCodeMatch ? errorCodeMatch[1] : "E0001";
    const queryParams = new URLSearchParams({
      username,
      name,
      external_email,
      remember: remember ? "1" : "0",
      error: errorCode,
    });
    redirect(`/signup?${queryParams.toString()}`, RedirectType.replace);
  });

  redirect("/signup/success", RedirectType.replace);
};
