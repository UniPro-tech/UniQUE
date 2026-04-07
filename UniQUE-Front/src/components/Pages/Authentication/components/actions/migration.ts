"use server";
import { RedirectType, redirect } from "next/navigation";
import { UserStatus } from "@/classes/types/User";
import { User } from "@/classes/User";
import { AuthenticationErrors } from "@/errors/AuthenticationErrors";

export const submitMigration = async (formData: FormData) => {
  const name = formData.get("name") as string;
  const email = formData.get("email") as string;
  const external_email = formData.get("external_email") as string;
  const password = formData.get("password") as string;
  const confirmPassword = formData.get("confirm_password") as string;
  const agreeTos = formData.get("agreeTos") === "on";

  if (password !== confirmPassword) {
    const queryParams = new URLSearchParams({
      email,
      name,
      external_email,
      agreeTos: agreeTos ? "1" : "0",
    });
    redirect(`/migrate?${queryParams.toString()}`, RedirectType.replace);
  }
  try {
    // emailの最初の.より前の部分をperiodとして使用
    const period = email.split("@")[0].split(".")[0];
    const username = email.split("@")[0].split(".").slice(1).join("."); // periodを除いた部分をusernameとして使用
    let internalEmail = email;
    if (period) {
      internalEmail = `${period.toUpperCase()}.${username}@uniproject.jp`;
    }

    const verifyRes = await fetch(
      `${process.env.GAS_MIGRATE_API_URL}?external_email=${encodeURIComponent(
        external_email,
      )}&internal_email=${encodeURIComponent(internalEmail)}`,
      { method: "GET" },
    );
    const verifyData = await verifyRes.json();
    if (!verifyRes.ok || verifyData.status !== 200) {
      console.log(
        `${process.env.GAS_MIGRATE_API_URL}?external_email=${encodeURIComponent(
          external_email,
        )}&internal_email=${encodeURIComponent(internalEmail)}`,
      );
      console.log("Migration verification failed:", verifyData);
      throw AuthenticationErrors.MigrationError;
    }

    const joinedAt = new Date(verifyData.joined_at);

    await User.create(
      {
        email,
        customId: username,
        externalEmail: external_email,
        affiliationPeriod: period.toUpperCase(),
        status: UserStatus.ACTIVE,
        profile: {
          displayName: name,
          joinedAt: joinedAt.toISOString(),
        },
      },
      password,
    );
  } catch (error) {
    const errorCodeMatch = (error as Error).message.match(/\[(.*?)\]/);
    const errorCode = errorCodeMatch ? errorCodeMatch[1] : "E0001";
    const queryParams = new URLSearchParams({
      name,
      email,
      external_email,
      agreeTos: agreeTos ? "1" : "0",
      error: errorCode,
    });
    redirect(`/migrate?${queryParams.toString()}`, RedirectType.replace);
  }

  redirect("/signin?migration=1", RedirectType.replace);
};
