"use server";

import { redirect } from "next/navigation";
import { Session } from "@/classes/Session";

export async function logoutAction() {
  try {
    (await Session.getCurrent())?.delete();
  } catch (error) {
    console.error("Logout failed:", error);
  } finally {
    redirect("/signin?signouted=true");
  }
}
