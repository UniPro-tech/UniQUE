"use client";

import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ErrorIcon from "@mui/icons-material/Error";
import {
  Alert,
  Box,
  Button,
  Card,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import type { EmailVerificationResponse } from "@/classes/User";

type VerifyState =
  | "loading"
  | "discord_required"
  | "success"
  | "error"
  | "invalid_code";

interface EmailVerifyClientProps {
  initialResult: EmailVerificationResponse | null;
  code: string | null;
  discordLinked?: boolean;
  discordError?: string;
}

export default function EmailVerifyClient({
  initialResult,
  code,
  discordLinked = false,
  discordError,
}: EmailVerifyClientProps) {
  const router = useRouter();

  const [state, setState] = useState<VerifyState>("loading");
  const [result, setResult] = useState<EmailVerificationResponse | null>(
    initialResult,
  );
  const [discordLinkingInProgress, setDiscordLinkingInProgress] =
    useState(false);

  // 初期化:メール検証結果の確認
  useEffect(() => {
    const initialize = async () => {
      if (!initialResult) {
        setState("invalid_code");
        return;
      }

      // Discord連携から戻ってきた場合
      if (discordLinked) {
        if ("valid" in initialResult && initialResult.valid) {
          if (
            "type" in initialResult &&
            initialResult.type === "discord_not_linked"
          ) {
            setResult(initialResult);
            setState("discord_required");
            return;
          }

          setResult(initialResult);
          setState("success");
          return;
        }

        // Discord連携後、初回だけメール検証APIを呼び出す
        try {
          if (code) {
            const verifyRes = await fetch(
              `/api/email-verify?code=${encodeURIComponent(code)}`,
            );
            if (verifyRes.ok) {
              const verifyData =
                (await verifyRes.json()) as EmailVerificationResponse;
              if (
                ("error" in verifyData &&
                  verifyData.error === "discord_not_linked") ||
                ("type" in verifyData &&
                  verifyData.type === "discord_not_linked")
              ) {
                setState("discord_required");
                setResult(verifyData as EmailVerificationResponse);
              } else if (verifyData.valid) {
                setResult(verifyData);
                setState("success");
              } else {
                setState("invalid_code");
              }
            } else {
              setState("error");
            }
          }
        } catch (error) {
          console.error("Failed to verify email after Discord link:", error);
          setState("error");
        }
        return;
      }

      // Discordエラーがある場合
      if (discordError) {
        if (discordError === "conflict") {
          // 既に連携済みの場合は、メール検証を再度試す
          if (code) {
            try {
              const verifyRes = await fetch(
                `/api/email-verify?code=${encodeURIComponent(code)}`,
              );
              if (verifyRes.ok) {
                const verifyData =
                  (await verifyRes.json()) as EmailVerificationResponse;
                if (
                  ("error" in verifyData &&
                    verifyData.error === "discord_not_linked") ||
                  ("type" in verifyData &&
                    verifyData.type === "discord_not_linked")
                ) {
                  setState("discord_required");
                  setResult(verifyData as EmailVerificationResponse);
                } else if (verifyData.valid) {
                  setResult(verifyData);
                  setState("success");
                } else {
                  setState("invalid_code");
                }
              } else {
                setState("error");
              }
            } catch (error) {
              console.error("Failed to verify email:", error);
              setState("error");
            }
          } else {
            setState("error");
          }
        } else {
          // その他のエラー
          setState("discord_required");
          setResult(initialResult);
        }
        return;
      }

      // `type` フィールドで Discord 未連携を示す場合
      if (
        "type" in initialResult &&
        initialResult.type === "discord_not_linked"
      ) {
        setState("discord_required");
        setResult(initialResult);
        return;
      }

      // エラーレスポンスの場合
      if ("error" in initialResult) {
        if (initialResult.error === "discord_not_linked") {
          // Discord未連携エラー
          setState("discord_required");
          setResult(initialResult as EmailVerificationResponse);
        } else {
          setState("invalid_code");
        }
        return;
      }

      // 失敗レスポンスの場合
      if (!initialResult.valid) {
        setState("invalid_code");
        return;
      }

      setResult(initialResult);
      setState("success");
    };

    initialize();
  }, [initialResult, discordLinked, discordError, code]);

  // Discord連携後に自動で次へ進む
  useEffect(() => {
    if (discordLinkingInProgress) {
      // 定期的にDiscord連携状態をチェック
      const interval = setInterval(async () => {
        try {
          const discordCheckRes = await fetch("/api/auth/check-discord-link");
          if (discordCheckRes.ok) {
            const data = await discordCheckRes.json();
            if (data.linked) {
              setDiscordLinkingInProgress(false);
              setState("success");
              clearInterval(interval);
            }
          }
        } catch (error) {
          console.error("Discord check error:", error);
        }
      }, 1000); // 1秒ごとにチェック

      return () => clearInterval(interval);
    }
  }, [discordLinkingInProgress]);

  useEffect(() => {
    if (state === "success" && result) {
      const redirectTimer = setTimeout(() => {
        const typeKey = (result.type ?? "") as string;
        if (typeKey === "signup" || typeKey === "registration") {
          router.push("/signup/success?email_verified=true");
        } else if (typeKey === "change" || typeKey === "email_change") {
          router.push(
            "/dashboard/settings?snack=メールアドレスの認証が完了しました。&variant=success",
          );
        } else if (typeKey === "migration") {
          router.push("/signin?migrated=true");
        } else {
          router.push("/signin");
        }
      }, 2000);

      return () => clearTimeout(redirectTimer);
    }
  }, [state, result, router]);

  const handleDiscordLink = () => {
    setDiscordLinkingInProgress(true);
    const params = new URLSearchParams({ from: "email_verify" });
    if (code) {
      params.set("code", code);
    }
    window.location.href = `/api/oauth/discord?${params.toString()}`;
  };

  const handleRetry = () => {
    router.push("/signin");
  };

  return (
    <Box
      sx={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
        padding: 2,
      }}
    >
      <Card
        sx={{
          width: "100%",
          maxWidth: 500,
          padding: 4,
          textAlign: "center",
        }}
      >
        {state === "loading" && (
          <Stack spacing={2} alignItems="center">
            <CircularProgress />
            <Typography variant="h6">認証中です...</Typography>
          </Stack>
        )}

        {state === "discord_required" && (
          <Stack spacing={2} alignItems="center">
            <Typography variant="h6">Discord連携が必要です</Typography>
            <Alert severity="info">
              メールアドレスの認証とユーザー登録を完了するには、Discord
              アカウントを連携してください。
            </Alert>
            {discordError && (
              <Alert severity="error">
                Discord連携エラー:{" "}
                {discordError === "conflict"
                  ? "このDiscordアカウントは既に使用されています。別のアカウントを使用するか、サポートにお問い合わせください。"
                  : "不明なエラーが発生しました。もう一度お試しください。"}
              </Alert>
            )}
            <Button
              variant="contained"
              onClick={handleDiscordLink}
              disabled={discordLinkingInProgress}
              sx={{ mt: 2 }}
            >
              {discordLinkingInProgress ? "連携中..." : "Discordを連携する"}
            </Button>
          </Stack>
        )}

        {state === "success" && (
          <Stack spacing={2} alignItems="center">
            <CheckCircleIcon sx={{ fontSize: 64, color: "success.main" }} />
            <Typography variant="h6">認証が完了しました!</Typography>
            <Typography sx={{ color: "text.secondary" }}>
              ページをリダイレクトしています...
            </Typography>
            <CircularProgress size={24} />
          </Stack>
        )}

        {state === "invalid_code" && (
          <Stack spacing={2} alignItems="center">
            <ErrorIcon sx={{ fontSize: 64, color: "error.main" }} />
            <Typography variant="h6">無効な認証コードです</Typography>
            <Typography sx={{ color: "text.secondary" }}>
              認証コードが無効か期限切れです。
              <br />
              もう一度サインアップからやり直してください。
            </Typography>
            <Button variant="contained" onClick={handleRetry} sx={{ mt: 2 }}>
              サインインページへ
            </Button>
          </Stack>
        )}

        {state === "error" && (
          <Stack spacing={2} alignItems="center">
            <ErrorIcon sx={{ fontSize: 64, color: "error.main" }} />
            <Typography variant="h6">エラーが発生しました</Typography>
            <Typography sx={{ color: "text.secondary" }}>
              認証処理中にエラーが発生しました。
              <br />
              もう一度お試しください。
            </Typography>
            <Button variant="contained" onClick={handleRetry} sx={{ mt: 2 }}>
              サインインページへ
            </Button>
          </Stack>
        )}
      </Card>
    </Box>
  );
}
