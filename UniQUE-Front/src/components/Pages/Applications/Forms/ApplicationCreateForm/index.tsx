"use client";

import {
  Cancel as CancelIcon,
  ContentCopy as ContentCopyIcon,
  Save as SaveIcon,
} from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  CircularProgress,
  IconButton,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { createApplication } from "./action";

interface ApplicationCreateFormProps {
  onCancel?: () => void;
  userId: string;
}

interface CreateApplicationFormData {
  name: string;
  clientSecret: string;
  description: string;
  websiteUrl: string;
  privacyPolicyUrl: string;
}

export default function ApplicationCreateForm({
  onCancel,
  userId,
}: ApplicationCreateFormProps) {
  const router = useRouter();

  const generateClientSecret = () => {
    // ランダムな64文字の英数字文字列を生成
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";
    for (let i = 0; i < 64; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  };

  const [formData, setFormData] = useState<CreateApplicationFormData>({
    name: "",
    clientSecret: generateClientSecret(),
    description: "",
    websiteUrl: "",
    privacyPolicyUrl: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copySuccess, setCopySuccess] = useState(false);

  const handleCopySecret = async () => {
    try {
      await navigator.clipboard.writeText(formData.clientSecret);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // バリデーション
    if (!formData.name.trim()) {
      setError("アプリケーション名を入力してください");
      setLoading(false);
      return;
    }

    try {
      const result = await createApplication({
        name: formData.name,
        clientSecret: formData.clientSecret,
        userId: userId,
        description: formData.description,
        websiteUrl: formData.websiteUrl,
        privacyPolicyUrl: formData.privacyPolicyUrl,
      });

      if (result.success) {
        // 成功したらアプリケーション一覧ページにリダイレクト
        router.push("/dashboard/applications");
      } else {
        setError(result.error || "作成に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      console.error("Application creation error:", {
        error: err,
        message: errorMessage,
        type: typeof err,
      });
      setError(errorMessage || "作成に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  const handleCancelClick = () => {
    if (onCancel) {
      onCancel();
    } else {
      router.push("/dashboard/applications");
    }
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        アプリケーション新規作成
      </Typography>

      <Box component="form" onSubmit={handleSubmit} noValidate>
        <Stack spacing={3}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          <TextField
            label="アプリケーション名"
            fullWidth
            required
            value={formData.name}
            onChange={(e) => {
              setFormData((prev) => ({ ...prev, name: e.target.value }));
              setError(null);
            }}
            helperText="アプリケーションの名前を入力してください"
            disabled={loading}
          />

          <Stack spacing={1}>
            <TextField
              label="クライアントシークレット"
              fullWidth
              required
              value={formData.clientSecret}
              helperText="OAuth2認証に使用するクライアントシークレット（自動生成）"
              disabled={true}
              type="password"
              slotProps={{
                input: {
                  readOnly: true,
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        onClick={handleCopySecret}
                        edge="end"
                        title={copySuccess ? "コピーしました!" : "コピー"}
                        disabled={loading}
                      >
                        <ContentCopyIcon
                          sx={{
                            color: copySuccess ? "success.main" : "inherit",
                          }}
                        />
                      </IconButton>
                    </InputAdornment>
                  ),
                },
              }}
            />
            <Button
              variant="outlined"
              size="small"
              onClick={() =>
                setFormData((prev) => ({
                  ...prev,
                  clientSecret: generateClientSecret(),
                }))
              }
              disabled={loading}
              sx={{ maxWidth: 150 }}
            >
              再生成
            </Button>
          </Stack>

          <TextField
            label="説明"
            fullWidth
            multiline
            rows={3}
            value={formData.description}
            onChange={(e) => {
              setFormData((prev) => ({ ...prev, description: e.target.value }));
              setError(null);
            }}
            helperText="アプリケーションの説明を入力してください（任意）"
            disabled={loading}
          />

          <TextField
            label="Webサイト URL"
            fullWidth
            type="url"
            value={formData.websiteUrl}
            onChange={(e) => {
              setFormData((prev) => ({ ...prev, websiteUrl: e.target.value }));
              setError(null);
            }}
            helperText="アプリケーションのWebサイトURL（任意）"
            disabled={loading}
          />

          <TextField
            label="プライバシーポリシー URL"
            fullWidth
            type="url"
            value={formData.privacyPolicyUrl}
            onChange={(e) => {
              setFormData((prev) => ({
                ...prev,
                privacyPolicyUrl: e.target.value,
              }));
              setError(null);
            }}
            helperText="プライバシーポリシーのURL（任意）"
            disabled={loading}
          />

          <Stack direction="row" spacing={2} justifyContent="flex-end">
            <Button
              variant="outlined"
              startIcon={<CancelIcon />}
              onClick={handleCancelClick}
              disabled={loading}
            >
              キャンセル
            </Button>

            <Button
              type="submit"
              variant="contained"
              startIcon={
                loading ? <CircularProgress size={20} /> : <SaveIcon />
              }
              disabled={loading}
            >
              {loading ? "作成中..." : "作成"}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Card>
  );
}
