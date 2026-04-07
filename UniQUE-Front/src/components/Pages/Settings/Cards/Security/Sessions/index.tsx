"use client";
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
  csrfToken,
}: {
  currentSessionId: string;
  sessions: SessionData[];
  csrfToken: string;
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
    <Stack spacing={2}>
      <Stack direction="row" alignItems="center" spacing={2}>
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
          }}
        >
          {displayedSessions.map((session) => {
            const ua = session.userAgent ? parseUA(session.userAgent) : null;
            const isDeleted = session.deletedAt !== null;
            return (
              <ListItem key={session.id} disablePadding>
                <Card
                  sx={{
                    p: 2,
                    width: "100%",
                    display: "flex",
                    flexDirection: "row",
                    alignItems: "center",
                    justifyContent: "space-between",
                    gap: 2,
                    opacity: isDeleted ? 0.6 : 1,
                  }}
                >
                  <Box>
                    <Stack direction="row" alignItems="center" spacing={1}>
                      <Typography variant="body1">
                        {ua
                          ? ua.browser !== "Unknown" || ua.os !== "Unknown"
                            ? `${ua.browser} - ${ua.os}`
                            : session.userAgent
                          : `セッション ${session.id.slice(0, 8)}`}
                      </Typography>
                      {session.id === activeSessionId && !isDeleted && (
                        <Chip
                          label="現在のセッション"
                          color="primary"
                          size="small"
                          variant="outlined"
                          sx={{ ml: 1 }}
                        />
                      )}
                      {isDeleted && (
                        <Chip
                          label="削除済み"
                          color="error"
                          size="small"
                          variant="outlined"
                          sx={{ ml: 1 }}
                        />
                      )}
                    </Stack>
                    {session.ipAddress && (
                      <Typography variant="body2" color="text.secondary">
                        IPアドレス: {session.ipAddress}
                      </Typography>
                    )}
                    {session.lastLoginAt && (
                      <Typography variant="body2" color="text.secondary">
                        最終ログイン:{" "}
                        {new Date(session.lastLoginAt).toLocaleString()}
                      </Typography>
                    )}
                    <Typography variant="body2" color="text.secondary">
                      作成日: {new Date(session.createdAt).toLocaleString()}
                    </Typography>
                    {session.expiresAt && (
                      <Typography variant="body2" color="text.secondary">
                        有効期限: {new Date(session.expiresAt).toLocaleString()}
                      </Typography>
                    )}
                  </Box>
                  {!isDeleted && (
                    <Box component="form" action={action}>
                      <input type="hidden" name="csrfToken" value={csrfToken} />
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
                        disabled={session.id === activeSessionId}
                        sx={{
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
