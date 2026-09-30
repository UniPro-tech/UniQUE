"use server";

import {
  type DiscordNotificationSettingsData,
  saveDiscordNotificationSettings,
} from "@/classes/DiscordNotificationSettings";

export async function updateDiscordNotificationSettings(
  settings: DiscordNotificationSettingsData,
): Promise<
  | { success: true; settings: DiscordNotificationSettingsData }
  | { success: false; error: string }
> {
  try {
    const updated = await saveDiscordNotificationSettings(settings);
    return { success: true, settings: updated };
  } catch (error) {
    console.error("Discord notification settings update failed:", error);
    return { success: false, error: "Discord通知設定の更新に失敗しました" };
  }
}
