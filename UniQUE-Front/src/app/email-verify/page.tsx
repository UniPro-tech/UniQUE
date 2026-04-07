import { redirect } from "next/navigation";
import { Session } from "@/classes/Session";
import EmailVerifyClient from "./EmailVerifyClient";

interface EmailVerifyPageProps {
  searchParams: Promise<{
    code?: string;
    discord_linked?: string;
    discord_error?: string;
  }>;
}

export default async function EmailVerifyPage({
  searchParams,
}: EmailVerifyPageProps) {
  const params = await searchParams;
  const session = await Session.getCurrent();
  if (session) await session.delete();
  const code = params.code;
  const discordLinked = params.discord_linked === "true";
  const discordError = params.discord_error;

  if (!code) {
    redirect("/signin");
  }

  // サーバーサイドでメール検証APIを叩く
  const baseUrl = "http://localhost:3000";
  let verifyResult = null;

  try {
    const response = await fetch(
      `${baseUrl}/api/email-verify?code=${encodeURIComponent(code)}`,
      {
        cache: "no-store",
      },
    );

    if (response.ok) {
      verifyResult = await response.json();
    } else {
      console.error(
        "Server-side email verification returned non-200:",
        response.status,
      );
    }
  } catch (error) {
    console.error("Server-side email verification error:", error);
  }

  return (
    <EmailVerifyClient
      initialResult={verifyResult}
      code={code}
      discordLinked={discordLinked}
      discordError={discordError}
    />
  );
}
