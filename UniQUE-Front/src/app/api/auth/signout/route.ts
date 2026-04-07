import { Session } from "@/classes/Session";

export const POST = async () => {
  (await Session.getCurrent())?.delete();
  return new Response(null, { status: 204 });
};
