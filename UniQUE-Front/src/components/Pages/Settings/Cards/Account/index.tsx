import type { UserData } from "@/classes/types/User";
import { generateCSRFToken } from "@/libs/csrf";
import AccountSettingsCardClient from "./Client";

export default async function AccountSettingsCard({
  user,
}: {
  user: UserData;
}) {
  const sid = user.id;
  const csrfToken = await generateCSRFToken(sid!);
  return <AccountSettingsCardClient user={user} csrfToken={csrfToken} />;
}
