"use client";
import { Card, Chip, Divider, Stack, Typography } from "@mui/material";
import React from "react";
import type { SessionData } from "@/classes/Session";
import type { UserData } from "@/classes/types/User";
import ForgotPassword from "@/components/Dialogs/ForgotPassword";
import PasswordSection from "./Password";
import SessionsSection from "./Sessions";
import TOTPSection from "./TOTP";

export default function SecuritySettingsCardClient({
  user,
  currentSessionId,
  sessions,
}: {
  user: UserData;
  currentSessionId: string;
  sessions: SessionData[];
}) {
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);

  const totpStatus = user.isTotpEnabled ? "有効" : "無効";
  const sessionCount = sessions.length;

  return (
    <>
      <Card
        variant="outlined"
        sx={{ display: "flex", p: 2, flexDirection: "column", gap: 2 }}
      >
        <Stack spacing={2}>
          <Stack>
            <Typography variant="h5" component={"h3"}>
              セキュリティ設定
            </Typography>
            <Typography variant="body2">
              パスワード、二段階認証、セッション管理をまとめて確認できます。
            </Typography>
          </Stack>

          <Stack
            direction="row"
            spacing={1}
            sx={{ alignItems: "center", flexWrap: "wrap" }}
          >
            <Chip
              label={`TOTP: ${totpStatus}`}
              color={user.isTotpEnabled ? "success" : "default"}
              size="small"
            />
            <Chip
              label={`ログイン中の端末: ${sessionCount}件`}
              variant="outlined"
              size="small"
            />
            <Chip label="パスワード保護中" variant="outlined" size="small" />
          </Stack>
        </Stack>

        <Divider />

        <PasswordSection user={user} handleClickOpen={handleClickOpen} />
        <TOTPSection user={user} />
        <SessionsSection
          currentSessionId={currentSessionId}
          sessions={sessions}
        />
      </Card>
      <ForgotPassword open={open} handleClose={handleClose} />
    </>
  );
}
