"use client";
import { Card, Stack, Typography } from "@mui/material";
import { SnackbarProvider } from "notistack";
import React from "react";
import type { SessionData } from "@/classes/Session";
import type { UserData } from "@/classes/types/User";
import ForgotPassword from "@/components/Dialogs/ForgotPassword";
import PasswordSection from "./Password";
import SessionsSection from "./Sessions";
import TOTPSection from "./TOTP";

export default function SecuritySettingsCardClient({
  user,
  csrfToken,
  currentSessionId,
  sessions,
}: {
  user: UserData;
  csrfToken: string;
  currentSessionId: string;
  sessions: SessionData[];
}) {
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Card
        variant="outlined"
        sx={{ display: "flex", p: 2, flexDirection: "column", gap: 2 }}
      >
        <Stack>
          <Typography variant="h5" component={"h3"}>
            セキュリティ設定
          </Typography>
          <Typography variant="body2">
            パスワードの変更や二段階認証の設定を行います。
          </Typography>
        </Stack>
        <PasswordSection
          user={user}
          csrfToken={csrfToken}
          handleClickOpen={handleClickOpen}
        />
        <TOTPSection user={user} csrfToken={csrfToken} />
        <SessionsSection
          currentSessionId={currentSessionId}
          sessions={sessions}
          csrfToken={csrfToken}
        />
      </Card>
      <ForgotPassword
        open={open}
        handleClose={handleClose}
        csrfToken={csrfToken}
      />
    </SnackbarProvider>
  );
}
