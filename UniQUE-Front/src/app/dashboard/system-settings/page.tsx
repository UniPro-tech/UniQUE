import { Stack, Typography } from "@mui/material";
import { getDiscordNotificationSettings } from "@/classes/DiscordNotificationSettings";
import DiscordNotificationSettingsForm from "@/components/Forms/DiscordNotificationSettingsForm";
import { PermissionBitsFields } from "@/constants/Permission";
import { requirePermission } from "@/libs/permissions";

export const metadata = {
  title: "全体設定",
  description: "UniQUE全体の設定を管理します。",
};

export default async function Page() {
  await requirePermission(PermissionBitsFields.CONFIG_UPDATE);
  const settings = await getDiscordNotificationSettings();

  return (
    <Stack spacing={4}>
      <Stack>
        <Typography variant="h4" component="h2">
          全体設定
        </Typography>
        <Typography variant="body1" color="text.secondary">
          UniQUE全体で利用する外部サービスの設定を管理します。
        </Typography>
      </Stack>
      <DiscordNotificationSettingsForm initialSettings={settings} />
    </Stack>
  );
}
