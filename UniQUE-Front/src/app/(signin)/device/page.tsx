import { RedirectType, redirect } from "next/navigation";
import { Session } from "@/classes/Session";
import DeviceFlowUserCodeInput from "@/components/Pages/Authorization/DeviceFlowUserCodeInput";

export default async function Page() {
  const session = await Session.getCurrent();

  if (!session) {
    const recirectpath = `/device`;
    redirect(
      `/signin?redirect=${encodeURIComponent(recirectpath)}`,
      RedirectType.replace,
    );
  }

  return <DeviceFlowUserCodeInput />;
}
