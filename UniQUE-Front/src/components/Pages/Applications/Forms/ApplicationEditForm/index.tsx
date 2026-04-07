"use client";

import {
  ContentCopy as ContentCopyIcon,
  Save as SaveIcon,
} from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  CircularProgress,
  Divider,
  FormControlLabel,
  IconButton,
  InputAdornment,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";
import type { ApplicationData } from "@/classes/Application";
import { changeAction } from "./action";

interface ApplicationEditFormProps {
  application: ApplicationData;
  owner?: {
    displayName: string;
    customId: string;
    email: string;
  };
}

interface UpdateApplicationFormData {
  name: string;
  description: string;
  websiteUrl: string;
  privacyPolicyUrl: string;
}

export default function ApplicationEditForm({
  application,
}: ApplicationEditFormProps) {
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

  const [formData, setFormData] = useState<UpdateApplicationFormData>({
    name: application.name || "",
    description: application.description || "",
    websiteUrl: application.websiteUrl || "",
    privacyPolicyUrl: application.privacyPolicyUrl || "",
  });

  const [clientSecret, setClientSecret] = useState<string | null>(null);
  const [regenerating, setRegenerating] = useState(false);
  const [secretSaved, setSecretSaved] = useState(false);
  const [clientIdCopySuccess, setClientIdCopySuccess] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  const handleChange = (
    field: keyof UpdateApplicationFormData,
    value: string,
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setError(null);
    setSuccess(false);
  };

  const handleCopySecret = async () => {
    if (!clientSecret) return;
    try {
      await navigator.clipboard.writeText(clientSecret);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  const handleCopyClientId = async () => {
    try {
      await navigator.clipboard.writeText(application.id);
      setClientIdCopySuccess(true);
      setTimeout(() => setClientIdCopySuccess(false), 2000);
    } catch (err) {
      console.error("Failed to copy client id:", err);
    }
  };

  const handleRegenerateSecret = () => {
    const newSecret = generateClientSecret();
    setClientSecret(newSecret);
    setCopySuccess(false);
    setSecretSaved(false);
    // 即時でサーバー側に保存する
    (async () => {
      setRegenerating(true);
      try {
        await changeAction({
          id: application.id,
          clientSecret: newSecret,
        });
        setSecretSaved(true);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "シークレットの再生成に失敗しました",
        );
        console.error(err);
      } finally {
        setRegenerating(false);
      }
    })();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(false);

    // バリデーション
    if (!formData.name.trim()) {
      setError("アプリケーション名を入力してください");
      setLoading(false);
      return;
    }

    try {
      const payload: Record<string, unknown> = {
        id: application.id,
        name: formData.name,
        description: formData.description,
        websiteUrl: formData.websiteUrl,
        privacyPolicyUrl: formData.privacyPolicyUrl,
      };
      if (clientSecret) {
        payload.clientSecret = clientSecret;
      }

      await changeAction(payload);

      setSuccess(true);
      // 保存後はシークレットをリセット
      setClientSecret(null);
      setCopySuccess(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新に失敗しました");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handlePublicClientToggle = async (
    e: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const isPublic = e.target.checked;
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      await changeAction({
        id: application.id,
        publicClient: isPublic,
      });
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新に失敗しました");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card sx={{ p: 4 }}>
      <Stack spacing={4}>
        {/* アプリケーション情報セクション */}
        <Box>
          <Typography variant="h6" gutterBottom>
            アプリケーション情報
          </Typography>
          <Stack
            spacing={2}
            sx={{ mt: 2 }}
            direction={{ xs: "column", sm: "row" }}
          >
            <Paper variant="outlined" sx={{ p: 2, flex: 1 }}>
              <Typography
                variant="caption"
                color="text.secondary"
                display="block"
              >
                アプリケーションID (クライアントID)
              </Typography>
              <Box sx={{ display: "flex", alignItems: "center", mt: 0.5 }}>
                <Typography
                  variant="body2"
                  sx={{
                    fontFamily: "monospace",
                    wordBreak: "break-all",
                    mr: 1,
                    flex: 1,
                  }}
                >
                  {application.id}
                </Typography>
                <IconButton
                  onClick={handleCopyClientId}
                  size="small"
                  title={clientIdCopySuccess ? "コピーしました!" : "コピー"}
                >
                  <ContentCopyIcon
                    sx={{
                      color: clientIdCopySuccess ? "success.main" : "inherit",
                    }}
                  />
                </IconButton>
              </Box>
            </Paper>
            <Paper variant="outlined" sx={{ p: 2, flex: 1 }}>
              <Typography
                variant="caption"
                color="text.secondary"
                display="block"
              >
                所有者
              </Typography>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: "monospace",
                  wordBreak: "break-all",
                  mt: 0.5,
                }}
              >
                {application.owner
                  ? `${application.owner.profile.displayName} (@${application.owner.customId})`
                  : "不明"}
              </Typography>
            </Paper>
          </Stack>
        </Box>

        <Divider />

        {/* クライアントシークレット管理セクション */}
        <Box>
          <Typography variant="h6" gutterBottom>
            クライアントシークレット管理
          </Typography>

          <Stack spacing={2} sx={{ mt: 2 }}>
            {clientSecret === null ? (
              <Button
                variant="outlined"
                color="warning"
                onClick={handleRegenerateSecret}
                disabled={loading || regenerating}
                sx={{ cursor: "pointer" }}
              >
                {regenerating ? (
                  <CircularProgress size={16} />
                ) : (
                  "シークレットを再生成"
                )}
              </Button>
            ) : (
              <Stack spacing={1}>
                <TextField
                  label="新しいクライアントシークレット"
                  fullWidth
                  required
                  value={clientSecret}
                  helperText="新しく生成されたシークレット（保存時に更新されます）"
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
                {secretSaved && (
                  <Alert severity="success">シークレットを更新しました</Alert>
                )}
                <Button
                  variant="outlined"
                  size="small"
                  onClick={handleRegenerateSecret}
                  disabled={loading || regenerating}
                  sx={{
                    cursor: loading || regenerating ? "progress" : "pointer",
                  }}
                >
                  {regenerating ? (
                    <CircularProgress size={16} />
                  ) : (
                    "別のシークレットを生成"
                  )}
                </Button>
              </Stack>
            )}
          </Stack>
          <FormControlLabel
            control={
              <Checkbox color="primary" onChange={handlePublicClientToggle} />
            }
            defaultChecked={application.publicClient}
            label="Public Client (PKCE必須)"
            sx={{ mt: 2 }}
          />
        </Box>

        <Divider />

        {/* 編集フォームセクション */}
        <Box>
          <Typography variant="h6" gutterBottom>
            基本情報編集
          </Typography>

          <Box
            component="form"
            onSubmit={handleSubmit}
            noValidate
            sx={{ mt: 2 }}
          >
            <Stack spacing={3}>
              {error && (
                <Alert severity="error" onClose={() => setError(null)}>
                  {error}
                </Alert>
              )}

              {success && (
                <Alert severity="success" onClose={() => setSuccess(false)}>
                  アプリケーションを更新しました
                </Alert>
              )}

              <TextField
                label="アプリケーション名"
                fullWidth
                required
                value={formData.name}
                onChange={(e) => handleChange("name", e.target.value)}
                helperText="アプリケーションの名前を入力してください"
                disabled={loading}
              />

              <TextField
                label="説明"
                fullWidth
                multiline
                rows={3}
                value={formData.description}
                onChange={(e) => handleChange("description", e.target.value)}
                helperText="アプリケーションの説明を入力してください（任意）"
                disabled={loading}
              />

              <TextField
                label="Webサイト URL"
                fullWidth
                type="url"
                value={formData.websiteUrl}
                onChange={(e) => handleChange("websiteUrl", e.target.value)}
                helperText="アプリケーションのWebサイトURL（任意）"
                disabled={loading}
              />

              <TextField
                label="プライバシーポリシー URL"
                fullWidth
                type="url"
                value={formData.privacyPolicyUrl}
                onChange={(e) =>
                  handleChange("privacyPolicyUrl", e.target.value)
                }
                helperText="プライバシーポリシーのURL（任意）"
                disabled={loading}
              />

              <Stack
                direction="row"
                spacing={2}
                justifyContent="flex-end"
                sx={{ pt: 2 }}
              >
                <Button
                  type="submit"
                  variant="contained"
                  startIcon={
                    loading ? <CircularProgress size={20} /> : <SaveIcon />
                  }
                  disabled={loading}
                >
                  {loading ? "保存中..." : "変更を保存"}
                </Button>
              </Stack>
            </Stack>
          </Box>
        </Box>
      </Stack>
    </Card>
  );
}
