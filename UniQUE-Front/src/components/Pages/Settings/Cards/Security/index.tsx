import { unauthorized } from "next/navigation";
import { Session } from "@/classes/Session";
import type { UserData } from "@/classes/types/User";
import { generateCSRFToken } from "@/libs/csrf";
import SecuritySettingsCardClient from "./Client";

export default async function SecuritySettingsCard({
  user,
}: {
  user: UserData;
}) {
  const uid = user.id!;
  const csrfToken = generateCSRFToken(uid);
  const session = await Session.getCurrent();
  if (!session) {
    unauthorized();
  }
  const sessions = (await Session.getByUserId(uid)).map((s) => s.toJson());
  return (
    <SecuritySettingsCardClient
      user={user}
      csrfToken={csrfToken}
      currentSessionId={session.id}
      sessions={sessions}
    />
  );
}
