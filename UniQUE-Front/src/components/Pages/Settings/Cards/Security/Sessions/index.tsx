"use client";
import {
  AccessTime as AccessTimeIcon,
  Devices as DevicesIcon,
  Logout as LogoutIcon,
  Public as PublicIcon,
} from "@mui/icons-material";
import {
  Box,
  Button,
  Card,
  Chip,
  Divider,
  List,
  ListItem,
  Stack,
  Typography,
} from "@mui/material";
import { enqueueSnackbar } from "notistack";
import { useActionState, useEffect } from "react";
import type { SessionData } from "@/classes/Session";
import { parseUA } from "@/libs/request";
import type { FormStatus } from "../../Base";
import { logoutSession } from "./action";

export default function SessionsSection({
  currentSessionId,
  sessions,
}: {
  currentSessionId: string;
  sessions: SessionData[];
}) {
  const [latestResult, action] = useActionState(logoutSession, {
    sessions: sessions as SessionData[],
    status: null as null | FormStatus,
  });

  useEffect(() => {
    if (latestResult.status) {
      enqueueSnackbar(latestResult.status.message, {
        variant: latestResult.status.status,
      });
    }
  }, [latestResult]);

  // 現在のセッションを判定: session_id が取得できればそれを使用、
  // できなければ lastLoginAt が最新のものを現在のセッションとみなす
  const getCurrentSessionId = () => {
    if (currentSessionId) return currentSessionId;
    if (latestResult.sessions.length === 0) return null;
    const sorted = [...latestResult.sessions].sort(
      (a, b) =>
        new Date(b.lastLoginAt || b.updatedAt).getTime() -
        new Date(a.lastLoginAt || a.updatedAt).getTime(),
    );
    return sorted[0].id;
  };

  const activeSessionId = getCurrentSessionId();

  // セッションを最大5件に制限
  const displayedSessions = latestResult.sessions.slice(0, 5);

  return (
    <Stack spacing={3}>
      {/* ヘッダー部分 */}
      <Stack
        direction="row"
        spacing={2}
        sx={{
          alignItems: "center",
        }}
      >
        <Typography variant="h6" component={"h4"}>
          セッション管理
        </Typography>
        <Divider sx={{ flexGrow: 1 }} />
      </Stack>

      {latestResult.sessions.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          現在アクティブなセッションはありません。
        </Typography>
      ) : (
        <List
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: 2,
            p: 0,
          }}
        >
          {displayedSessions.map((session) => {
            const ua = session.userAgent ? parseUA(session.userAgent) : null;
            const isDeleted = session.deletedAt !== null;
            const isActive = session.id === activeSessionId && !isDeleted;

            return (
              <ListItem key={session.id} disablePadding>
                <Card
                  variant="outlined"
                  sx={{
                    p: 2.5,
                    width: "100%",
                    display: "flex",
                    flexDirection: { xs: "column", sm: "row" },
                    alignItems: { xs: "flex-start", sm: "center" },
                    justifyContent: "space-between",
                    gap: 2,
                    opacity: isDeleted ? 0.6 : 1,
                    transition: "box-shadow 0.2s",
                    "&:hover": {
                      boxShadow: isDeleted ? "none" : 1,
                    },
                  }}
                >
                  <Stack
                    direction="row"
                    spacing={2}
                    sx={{ flexGrow: 1, alignItems: "flex-start" }}
                  >
                    {/* アイコン部分 */}
                    <Box
                      sx={{
                        p: 1.5,
                        borderRadius: 2,
                        bgcolor: isActive ? "primary.50" : "action.hover",
                        color: isActive ? "primary.main" : "text.secondary",
                        display: "flex",
                      }}
                    >
                      <DevicesIcon />
                    </Box>

                    {/* 情報部分 */}
                    <Stack spacing={1} sx={{ flexGrow: 1 }}>
                      <Stack
                        direction={{ xs: "column", sm: "row" }}
                        spacing={1}
                        useFlexGap
                        sx={{
                          alignItems: { xs: "flex-start", sm: "center" },
                          flexGap: 1,
                        }}
                      >
                        <Typography
                          variant="subtitle1"
                          sx={{ fontWeight: "bold", lineHeight: 1.2 }}
                        >
                          {ua
                            ? ua.browser !== "Unknown" || ua.os !== "Unknown"
                              ? `${ua.browser} - ${ua.os}`
                              : session.userAgent
                            : `セッション ${session.id.slice(0, 8)}`}
                        </Typography>

                        {isActive && (
                          <Chip
                            label="現在のセッション"
                            color="primary"
                            size="small"
                          />
                        )}
                        {isDeleted && (
                          <Chip
                            label="削除済み"
                            color="error"
                            size="small"
                            variant="outlined"
                          />
                        )}
                      </Stack>

                      {/* メタデータ群 */}
                      <Stack
                        direction={{ xs: "column", sm: "row" }}
                        spacing={2}
                        useFlexGap
                        sx={{
                          color: "text.secondary",
                          mt: 0.5,
                          flexWrap: "wrap",
                        }}
                      >
                        {session.ipAddress && (
                          <Stack
                            direction={{ xs: "row" }}
                            spacing={0.5}
                            sx={{
                              alignItems: { xs: "flex-start", sm: "center" },
                            }}
                          >
                            <Stack
                              direction="row"
                              spacing={0.5}
                              sx={{ alignItems: "center" }}
                            >
                              <PublicIcon fontSize="small" color="inherit" />
                              <Typography variant="caption">IP:</Typography>
                            </Stack>
                            <Typography variant="caption">
                              {session.ipAddress}
                            </Typography>
                          </Stack>
                        )}
                        {session.lastLoginAt && (
                          <Stack
                            direction={{ xs: "column", sm: "row" }}
                            spacing={0.5}
                            sx={{
                              alignItems: { xs: "flex-start", sm: "center" },
                            }}
                          >
                            <Stack
                              direction="row"
                              spacing={0.5}
                              sx={{ alignItems: "center" }}
                            >
                              <AccessTimeIcon
                                fontSize="small"
                                color="inherit"
                              />
                              <Typography variant="caption">
                                最終ログイン:
                              </Typography>
                            </Stack>
                            <Typography variant="caption">
                              {new Date(session.lastLoginAt).toLocaleString()}
                            </Typography>
                          </Stack>
                        )}
                      </Stack>
                    </Stack>
                  </Stack>

                  {/* アクション部分 */}
                  {!isDeleted && (
                    <Box
                      component="form"
                      action={action}
                      sx={{
                        width: { xs: "100%", sm: "auto" },
                        display: "flex",
                        justifyContent: "flex-end",
                      }}
                    >
                      <input
                        type="hidden"
                        name="sessionId"
                        value={session.id}
                      />
                      <Button
                        type="submit"
                        variant="outlined"
                        size="small"
                        color="error"
                        startIcon={<LogoutIcon />}
                        disabled={session.id === activeSessionId}
                        sx={{
                          whiteSpace: "nowrap",
                          cursor:
                            session.id === activeSessionId
                              ? "not-allowed"
                              : "pointer",
                        }}
                      >
                        ログアウト
                      </Button>
                    </Box>
                  )}
                </Card>
              </ListItem>
            );
          })}
        </List>
      )}
    </Stack>
  );
}
