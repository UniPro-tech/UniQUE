"use client";

import AppsIcon from "@mui/icons-material/Apps";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Box,
  Button,
  Card,
  Chip,
  Divider,
  Link,
  Stack,
  Typography,
} from "@mui/material";
import { enqueueSnackbar } from "notistack";
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
  applicationTermsUrl?: string;
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
    <Card
      variant="outlined"
      sx={{
        p: 3,
        display: "flex",
        flexDirection: "column",
        gap: 2.5,
      }}
    >
      {/* ヘッダー部分 */}
      <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
        <Typography variant="h6" sx={{ component: "h3" }}>
          アプリ連携（OAuth同意）
        </Typography>
        <Divider sx={{ flexGrow: 1, ml: 1 }} />
      </Stack>

      <Typography variant="body2" color="text.secondary" sx={{ mt: -0.5 }}>
        アクセスを許可したアプリケーションの一覧です。取り消すと、そのアプリからのアクセスが無効になります。
      </Typography>

      {consents.length === 0 ? (
        <Box
          sx={{
            p: 3,
            bgcolor: "action.hover",
            borderRadius: 2,
            textAlign: "center",
          }}
        >
          <Typography variant="body2" color="text.secondary">
            現在、連携しているアプリケーションはありません。
          </Typography>
        </Box>
      ) : (
        <Box>
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
              <Accordion
                key={consent.id}
                disableGutters
                elevation={0}
                sx={{
                  border: "1px solid",
                  borderColor: "divider",
                  borderRadius: 1.5,
                  overflow: "hidden",
                  "&:before": { display: "none" }, // デフォルトの怪しい線を消去
                  "&:not(:last-child)": { mb: 1.5 }, // アプリ同士の適度な隙間
                }}
              >
                <AccordionSummary
                  expandIcon={<ExpandMoreIcon />}
                  sx={{
                    "&:hover": { bgcolor: "action.hover" },
                    px: { xs: 2, sm: 3 },
                  }}
                >
                  <Stack spacing={0.5} sx={{ minWidth: 0 }}>
                    <Typography
                      variant="subtitle2"
                      sx={{ fontWeight: "bold" }}
                      noWrap
                    >
                      {appLabel}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" noWrap>
                      {consent.applicationDescription ||
                        "クリックして詳細を表示"}
                    </Typography>
                  </Stack>
                </AccordionSummary>

                <AccordionDetails sx={{ pt: 0, pb: 2, px: { xs: 2, sm: 3 } }}>
                  <Stack spacing={2}>
                    <Divider />

                    {/* リンク群 */}
                    <Stack spacing={1}>
                      {[
                        {
                          label: "ウェブサイト",
                          url: consent.applicationWebsiteUrl,
                        },
                        { label: "利用規約", url: consent.applicationTermsUrl },
                        {
                          label: "プライバシーポリシー",
                          url: consent.applicationPrivacyPolicyUrl,
                        },
                      ].map((link) =>
                        link.url ? (
                          <Typography
                            key={link.label}
                            variant="body2"
                            sx={{
                              display: "flex",
                              alignItems: { xs: "flex-start", sm: "center" },
                              gap: 1,
                              flexDirection: { xs: "column", sm: "row" },
                            }}
                          >
                            <Typography
                              component="span"
                              variant="body2"
                              color="text.secondary"
                              sx={{ minWidth: 145 }}
                            >
                              {link.label}:
                            </Typography>
                            <Link
                              href={link.url}
                              target="_blank"
                              rel="noopener noreferrer"
                              underline="hover"
                              sx={{
                                display: "flex",
                                alignItems: "center",
                                gap: 0.5,
                              }}
                            >
                              リンクを開く{" "}
                              <OpenInNewIcon sx={{ fontSize: 16 }} />
                            </Link>
                          </Typography>
                        ) : null,
                      )}
                    </Stack>

                    {/* スコープ群 */}
                    {scopes.length > 0 && (
                      <Stack spacing={1}>
                        <Typography variant="caption" color="text.secondary">
                          許可された権限:
                        </Typography>
                        <Stack
                          direction="row"
                          spacing={0.5}
                          useFlexGap
                          sx={{ flexWrap: "wrap" }}
                        >
                          {scopes.map((s) => (
                            <Chip
                              key={s}
                              label={s}
                              size="small"
                              variant="outlined"
                            />
                          ))}
                        </Stack>
                      </Stack>
                    )}

                    {/* アクションと日付 */}
                    <Stack
                      direction={{ xs: "column", sm: "row" }}
                      sx={{
                        pt: 1,
                        justifyContent: "space-between",
                        alignItems: { xs: "flex-start", sm: "center" },
                        gap: 1,
                      }}
                    >
                      <Typography variant="caption" color="text.secondary">
                        {consent.createdAt
                          ? `許可日: ${new Date(consent.createdAt).toLocaleDateString()}`
                          : ""}
                      </Typography>
                      <Button
                        variant="outlined"
                        color="error"
                        size="small"
                        disabled={revoking === consent.id}
                        onClick={() => handleRevoke(consent)}
                        disableElevation
                      >
                        {revoking === consent.id
                          ? "取り消し中..."
                          : "連携を取り消す"}
                      </Button>
                    </Stack>
                  </Stack>
                </AccordionDetails>
              </Accordion>
            );
          })}
        </Box>
      )}
    </Card>
  );
}
