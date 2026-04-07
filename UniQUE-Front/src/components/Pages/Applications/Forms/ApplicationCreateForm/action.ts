"use server";

import { Application } from "@/classes/Application";

export interface CreateApplicationFormData {
  name: string;
  clientSecret: string;
  userId: string;
  description: string;
  websiteUrl: string;
  privacyPolicyUrl: string;
}

export async function createApplication(
  data: CreateApplicationFormData,
): Promise<{ success: boolean; error?: string; applicationId?: string }> {
  try {
    if (!data.name.trim()) {
      return {
        success: false,
        error: "アプリケーション名を入力してください",
      };
    }

    if (!data.clientSecret.trim()) {
      return {
        success: false,
        error: "クライアントシークレットを生成してください",
      };
    }

    const application = await Application.create({
      name: data.name,
      clientSecret: data.clientSecret,
      description: data.description || null,
      websiteUrl: data.websiteUrl || null,
      privacyPolicyUrl: data.privacyPolicyUrl || null,
      userId: data.userId,
    });

    return { success: true, applicationId: application.id };
  } catch (error) {
    console.error("Application creation failed:", error);
    return {
      success: false,
      error:
        error instanceof Error
          ? error.message
          : "アプリケーションの作成に失敗しました",
    };
  }
}
