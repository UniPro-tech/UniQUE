import type { UserData } from "@/classes/types/User";
import AccountSettingsCardClient from "./Client";

export default async function AccountSettingsCard({
  user,
}: {
  user: UserData;
}) {
  return <AccountSettingsCardClient user={user} />;
}
