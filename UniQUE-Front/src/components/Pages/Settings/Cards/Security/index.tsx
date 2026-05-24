import { unauthorized } from "next/navigation";
import { Session } from "@/classes/Session";
import type { UserData } from "@/classes/types/User";
import SecuritySettingsCardClient from "./Client";

export default async function SecuritySettingsCard({
  user,
}: {
  user: UserData;
}) {
  const uid = user.id;
  const session = await Session.getCurrent();
  if (!session) {
    unauthorized();
  }
  const sessions = (await Session.getByUserId(uid)).map((s) => s.toJson());
  return (
    <SecuritySettingsCardClient
      user={user}
      currentSessionId={session.id}
      sessions={sessions}
    />
  );
}
