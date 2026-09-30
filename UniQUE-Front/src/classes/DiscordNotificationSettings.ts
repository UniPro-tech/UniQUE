import { apiGet, apiPut } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";

export interface DiscordNotificationSettingsData {
  channelId: string;
  notifyAnnouncements: boolean;
  notifyRegistrationRequests: boolean;
  notifyMigrations: boolean;
}

const endpoint = "/settings/discord-notifications";

export async function getDiscordNotificationSettings(): Promise<DiscordNotificationSettingsData> {
  const response = await apiGet(endpoint);
  if (!response.ok) {
    throw new Error(
      `Failed to fetch Discord notification settings: ${response.statusText}`,
    );
  }
  return toCamelcase<DiscordNotificationSettingsData>(await response.json());
}

export async function saveDiscordNotificationSettings(
  settings: DiscordNotificationSettingsData,
): Promise<DiscordNotificationSettingsData> {
  const response = await apiPut(endpoint, settings);
  if (!response.ok) {
    throw new Error(
      `Failed to update Discord notification settings: ${response.statusText}`,
    );
  }
  return toCamelcase<DiscordNotificationSettingsData>(await response.json());
}
