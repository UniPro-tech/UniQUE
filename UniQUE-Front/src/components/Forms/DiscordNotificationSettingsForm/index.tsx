"use client";

import NotificationsActiveIcon from "@mui/icons-material/NotificationsActive";
import SaveIcon from "@mui/icons-material/Save";
import {
  Alert,
  Button,
  Card,
  CircularProgress,
  FormControlLabel,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";
import type { DiscordNotificationSettingsData } from "@/classes/DiscordNotificationSettings";
import { updateDiscordNotificationSettings } from "./action";

const discordChannelIDPattern = /^[0-9]{17,20}$/;

export default function DiscordNotificationSettingsForm({
  initialSettings,
}: {
  initialSettings: DiscordNotificationSettingsData;
}) {
  const [settings, setSettings] =
    useState<DiscordNotificationSettingsData>(initialSettings);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleToggle = (
    field: keyof Pick<
      DiscordNotificationSettingsData,
      "notifyAnnouncements" | "notifyRegistrationRequests" | "notifyMigrations"
    >,
    checked: boolean,
  ) => {
    setSettings((current) => ({ ...current, [field]: checked }));
    setError(null);
    setSuccess(false);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setError(null);
    setSuccess(false);

    const channelID = settings.channelId.trim();
    if (channelID !== "" && !discordChannelIDPattern.test(channelID)) {
      setError("チャンネルIDは17〜20桁の数字で入力してください");
      return;
    }

    setLoading(true);
    const result = await updateDiscordNotificationSettings({
      ...settings,
      channelId: channelID,
    });
    if (result.success) {
      setSettings(result.settings);
      setSuccess(true);
    } else {
      setError(result.error);
    }
    setLoading(false);
  };

  return (
    <Card
      component="form"
      onSubmit={handleSubmit}
      variant="outlined"
      sx={{ p: 3 }}
    >
      <Stack spacing={3}>
        <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
          <NotificationsActiveIcon color="primary" />
          <Typography variant="h6">Discord通知</Typography>
        </Stack>

        <Typography variant="body2" color="text.secondary">
          通知先チャンネルと、Discordへ通知するイベントを設定します。通知には個人情報やアナウンス本文を含めません。
        </Typography>

        {error && <Alert severity="error">{error}</Alert>}
        {success && <Alert severity="success">設定を保存しました</Alert>}

        <TextField
          label="通知先チャンネルID"
          value={settings.channelId}
          onChange={(event) => {
            setSettings((current) => ({
              ...current,
              channelId: event.target.value,
            }));
            setError(null);
            setSuccess(false);
          }}
          inputMode="numeric"
          helperText="DiscordのチャンネルIDを入力します。空欄で通知を停止します。"
          disabled={loading}
          fullWidth
        />

        <Stack>
          <FormControlLabel
            control={
              <Switch
                checked={settings.notifyAnnouncements}
                onChange={(event) =>
                  handleToggle("notifyAnnouncements", event.target.checked)
                }
                disabled={loading}
              />
            }
            label="アナウンスが追加された時"
          />
          <FormControlLabel
            control={
              <Switch
                checked={settings.notifyRegistrationRequests}
                onChange={(event) =>
                  handleToggle(
                    "notifyRegistrationRequests",
                    event.target.checked,
                  )
                }
                disabled={loading}
              />
            }
            label="メンバー登録申請が届いた時"
          />
          <FormControlLabel
            control={
              <Switch
                checked={settings.notifyMigrations}
                onChange={(event) =>
                  handleToggle("notifyMigrations", event.target.checked)
                }
                disabled={loading}
              />
            }
            label="アカウント移行によりメンバーが追加された時"
          />
        </Stack>

        <Button
          type="submit"
          variant="contained"
          startIcon={loading ? <CircularProgress size={20} /> : <SaveIcon />}
          disabled={loading}
          sx={{ alignSelf: "flex-end" }}
        >
          {loading ? "保存中..." : "設定を保存"}
        </Button>
      </Stack>
    </Card>
  );
}
