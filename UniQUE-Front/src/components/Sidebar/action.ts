"use server";

import { redirect } from "next/navigation";
import { Session } from "@/classes/Session";
import { logger } from "@/libs/logger";

export async function logoutAction() {
  try {
    (await Session.getCurrent())?.delete();
  } catch (error) {
    logger.error("Logout failed:", error);
  } finally {
    redirect("/signin?signouted=true");
  }
}
