"use client";

import {
  Box,
  Button,
  Card,
  Chip,
  List,
  ListItem,
  Stack,
  Typography,
} from "@mui/material";
import { enqueueSnackbar, SnackbarProvider } from "notistack";
import React from "react";

export interface ConsentDTO {
  id: string;
  clientId?: string;
  applicationId?: string;
  userId?: string;
  scope?: string;
  createdAt?: string;
  /** フロントで解決したアプリ名 */
  applicationName?: string;
  applicationDescription?: string;
  applicationWebsiteUrl?: string;
  applicationPrivacyPolicyUrl?: string;
}

async function clientRevokeConsent(consentId: string): Promise<boolean> {
  const authApiUrl = process.env.NEXT_PUBLIC_AUTH_API_URL;
  const res = await fetch(
    `${authApiUrl}/internal/consents/${encodeURIComponent(consentId)}`,
    {
      method: "DELETE",
      credentials: "include",
    },
  );
  return res.ok;
}

export default function ConsentSettingsCardClient({
  consents: initialConsents,
  revokeConsent,
}: {
  consents: ConsentDTO[];
  revokeConsent?: (id: string) => Promise<boolean>;
}) {
  const [consents, setConsents] = React.useState(initialConsents);
  const [revoking, setRevoking] = React.useState<string | null>(null);

  // アプリ名の解決はサーバーサイドで行われる想定なので、ここでは何もしない

  const handleRevoke = async (consent: ConsentDTO) => {
    setRevoking(consent.id);
    try {
      const ok = await (revokeConsent
        ? revokeConsent(consent.id)
        : clientRevokeConsent(consent.id));
      if (ok) {
        setConsents((prev) => prev.filter((c) => c.id !== consent.id));
        enqueueSnackbar("同意を取り消しました。", { variant: "success" });
      } else {
        enqueueSnackbar("同意の取り消しに失敗しました。", {
          variant: "error",
        });
      }
    } catch {
      enqueueSnackbar("同意の取り消し中にエラーが発生しました。", {
        variant: "error",
      });
    } finally {
      setRevoking(null);
    }
  };

  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Card
        variant="outlined"
        sx={{ display: "flex", p: 2, flexDirection: "column", gap: 2 }}
      >
        <Stack>
          <Typography variant="h5" component="h3">
            アプリ連携（OAuth同意）
          </Typography>
          <Typography variant="body2">
            アクセスを許可したアプリケーションの一覧です。取り消すと、そのアプリからのアクセスが無効になります。
          </Typography>
        </Stack>

        {consents.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            現在、同意しているアプリケーションはありません。
          </Typography>
        ) : (
          <List sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
            {consents.map((consent) => {
              const appLabel =
                consent.applicationName ||
                consent.clientId ||
                consent.applicationId ||
                consent.id;
              const scopes = consent.scope
                ? consent.scope.split(" ").filter(Boolean)
                : [];
              return (
                <ListItem key={consent.id} disablePadding>
                  <Card
                    sx={{
                      p: 2,
                      width: "100%",
                      display: "flex",
                      flexDirection: "row",
                      alignItems: "center",
                      justifyContent: "space-between",
                      gap: 2,
                    }}
                  >
                    <Box
                      sx={{
                        flex: 1,
                        minWidth: 0,
                        display: "flex",
                        flexDirection: "column",
                        gap: 1,
                      }}
                    >
                      <Typography variant="subtitle1" noWrap>
                        {appLabel}
                      </Typography>
                      {consent.applicationDescription && (
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          noWrap
                        >
                          {consent.applicationDescription}
                        </Typography>
                      )}
                      {consent.applicationWebsiteUrl && (
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          noWrap
                        >
                          ウェブサイト:{" "}
                          <a
                            href={consent.applicationWebsiteUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {consent.applicationWebsiteUrl}
                          </a>
                        </Typography>
                      )}
                      {consent.applicationPrivacyPolicyUrl && (
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          noWrap
                        >
                          プライバシーポリシー:{" "}
                          <a
                            href={consent.applicationPrivacyPolicyUrl}
                            target="_blank"
                            rel="noopener noreferrer"
                          >
                            {consent.applicationPrivacyPolicyUrl}
                          </a>
                        </Typography>
                      )}
                      {scopes.length > 0 && (
                        <Stack direction="row" spacing={0.5} flexWrap="wrap">
                          {scopes.map((s) => (
                            <Chip key={s} label={s} size="small" />
                          ))}
                        </Stack>
                      )}
                      {consent.createdAt && (
                        <Typography variant="caption" color="text.secondary">
                          許可日:{" "}
                          {new Date(consent.createdAt).toLocaleDateString()}
                        </Typography>
                      )}
                    </Box>
                    <Button
                      variant="outlined"
                      color="error"
                      size="small"
                      disabled={revoking === consent.id}
                      onClick={() => handleRevoke(consent)}
                    >
                      {revoking === consent.id ? "取り消し中..." : "取り消す"}
                    </Button>
                  </Card>
                </ListItem>
              );
            })}
          </List>
        )}
      </Card>
    </SnackbarProvider>
  );
}
