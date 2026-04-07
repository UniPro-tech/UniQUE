"use server";

import { redirect } from "next/navigation";
import type { UserStatus } from "@/classes/types/User";
import { User } from "@/classes/User";

export interface CreateMemberFormData {
  customId: string;
  email: string;
  password: string;
  displayName?: string;
  status?: UserStatus;
  externalEmail?: string;
  affiliationPeriod?: string;
}

export async function createMember(
  data: CreateMemberFormData,
): Promise<{ success: boolean; error?: string; userId?: string }> {
  try {
    const user = await User.create(
      {
        customId: data.customId,
        email: data.email,
        externalEmail: data.externalEmail || null,
        affiliationPeriod: data.affiliationPeriod,
        status: data.status,
        profile: {
          displayName: data.displayName,
        },
      },
      data.password,
    );
    return { success: true, userId: user.id };
  } catch (err) {
    console.error("Member creation failed:", err);
    return { success: false, error: "ユーザーの作成に失敗しました" };
  }
}

export async function redirectToMemberList() {
  redirect("/dashboard/members");
}
