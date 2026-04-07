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
  TextField,
  Typography,
} from "@mui/material";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { confirmPasswordReset, passwordResetRequest } from "./action";

type PasswordResetState =
  | "request"
  | "loading"
  | "confirm"
  | "success"
  | "error"
  | "invalid_code";

export default function PasswordResetClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const code = searchParams.get("code");

  const [state, setState] = useState<PasswordResetState>(
    code ? "loading" : "request",
  );
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (code) {
      setState("confirm");
    }
  }, [code]);

  const handleRequestPasswordReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      await passwordResetRequest(email);
      setState("success");
    } catch (_err) {
      setError("エラーが発生しました。");
      setState("error");
    } finally {
      setIsLoading(false);
    }
  };

  const handleConfirmPasswordReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    if (password !== confirmPassword) {
      setError("パスワードと確認用パスワードが一致しません。");
      setIsLoading(false);
      return;
    }

    if (password.length < 8) {
      setError("パスワードは8文字以上である必要があります。");
      setIsLoading(false);
      return;
    }

    try {
      if (!code) {
        throw new Error("Invalid code");
      }

      await confirmPasswordReset(code, password);
      setState("success");
    } catch (_err) {
      setError("エラーが発生しました。");
      setState("error");
    } finally {
      setIsLoading(false);
    }
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
        {state === "request" && (
          <Stack spacing={2} alignItems="center">
            <Typography variant="h6">パスワードをリセット</Typography>
            <form
              onSubmit={handleRequestPasswordReset}
              style={{ width: "100%" }}
            >
              <Stack spacing={2}>
                <TextField
                  fullWidth
                  label="メールアドレス"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isLoading}
                  placeholder="example@example.com"
                />
                <Button
                  variant="contained"
                  type="submit"
                  fullWidth
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <CircularProgress size={24} />
                  ) : (
                    "パスワードリセットメールを送信"
                  )}
                </Button>
              </Stack>
            </form>
          </Stack>
        )}

        {state === "confirm" && code && (
          <Stack spacing={2} alignItems="center">
            <Typography variant="h6">新しいパスワードを設定</Typography>
            <form
              onSubmit={handleConfirmPasswordReset}
              style={{ width: "100%" }}
            >
              <Stack spacing={2}>
                {error && <Alert severity="error">{error}</Alert>}
                <TextField
                  fullWidth
                  label="新しいパスワード"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isLoading}
                  helperText="8文字以上である必要があります"
                />
                <TextField
                  fullWidth
                  label="パスワード（確認）"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                  disabled={isLoading}
                />
                <Button
                  variant="contained"
                  type="submit"
                  fullWidth
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <CircularProgress size={24} />
                  ) : (
                    "パスワードをリセット"
                  )}
                </Button>
              </Stack>
            </form>
          </Stack>
        )}

        {state === "success" && (
          <Stack spacing={2} alignItems="center">
            <CheckCircleIcon sx={{ fontSize: 64, color: "success.main" }} />
            <Typography variant="h6">パスワードをリセットしました！</Typography>
            <Typography color="textSecondary">
              新しいパスワードでサインインしてください。
            </Typography>
            <Button
              variant="contained"
              onClick={() => router.push("/signin")}
              sx={{ mt: 2 }}
            >
              サインインページへ
            </Button>
          </Stack>
        )}

        {state === "invalid_code" && (
          <Stack spacing={2} alignItems="center">
            <ErrorIcon sx={{ fontSize: 64, color: "error.main" }} />
            <Typography variant="h6">無効な認証コードです</Typography>
            <Typography color="textSecondary">
              認証コードが無効か期限切れです。
              <br />
              もう一度パスワードリセットをリクエストしてください。
            </Typography>
            <Button variant="contained" onClick={handleRetry} sx={{ mt: 2 }}>
              パスワードリセットをリクエスト
            </Button>
          </Stack>
        )}

        {state === "error" && (
          <Stack spacing={2} alignItems="center">
            <ErrorIcon sx={{ fontSize: 64, color: "error.main" }} />
            <Typography variant="h6">エラーが発生しました</Typography>
            <Typography color="textSecondary">
              {error || "処理中にエラーが発生しました。"}
              <br />
              もう一度お試しください。
            </Typography>
            <Button
              variant="contained"
              onClick={() => setState("request")}
              sx={{ mt: 2 }}
            >
              もう一度試す
            </Button>
          </Stack>
        )}

        {state === "loading" && (
          <Stack spacing={2} alignItems="center">
            <CircularProgress />
            <Typography variant="h6">処理中です...</Typography>
          </Stack>
        )}
      </Card>
    </Box>
  );
}
