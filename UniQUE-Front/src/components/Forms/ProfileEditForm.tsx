"use client";

import { Cancel as CancelIcon, Save as SaveIcon } from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  CircularProgress,
  FormControlLabel,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { useState } from "react";
import type { ProfileData } from "@/classes/Profile";
import {
  type UpdateProfileData,
  updateProfile,
} from "@/components/Cards/Profile/action";

interface ProfileEditFormProps {
  profile?: ProfileData;
  onCancel?: () => void;
  onSuccess?: () => void;
  setProfile?: (data: ProfileData) => void;
}

export default function ProfileEditForm({
  profile,
  onCancel,
  onSuccess,
  setProfile,
}: ProfileEditFormProps) {
  const { enqueueSnackbar } = useSnackbar();
  const [formData, setFormData] = useState<UpdateProfileData>({
    displayName: profile?.displayName || "",
    bio: profile?.bio || "",
    websiteUrl: profile?.websiteUrl || "",
    twitterHandle: profile?.twitterHandle || "",
    birthdateVisible: profile?.birthdateVisible ?? false,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleChange = (
    field: keyof UpdateProfileData,
    value: string | boolean,
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setError(null);
    setSuccess(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const result = await updateProfile(formData);

      if (result.success) {
        enqueueSnackbar(result.message || "プロフィールを更新しました", {
          variant: "success",
        });
        if (setProfile) setProfile(result.profile!);
        if (onSuccess) onSuccess();
      } else {
        const msg = result.error || "更新に失敗しました";
        setError(msg);
        enqueueSnackbar(msg, { variant: "error" });
      }
    } catch (err) {
      setError("予期しないエラーが発生しました");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        プロフィール編集
      </Typography>

      <Box component="form" onSubmit={handleSubmit} noValidate>
        <Stack spacing={3}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          {success && (
            <Alert severity="success" onClose={() => setSuccess(false)}>
              プロフィールを更新しました
            </Alert>
          )}

          <TextField
            label="表示名"
            fullWidth
            value={formData.displayName}
            onChange={(e) => handleChange("displayName", e.target.value)}
            helperText="あなたの表示名を入力してください"
            disabled={loading}
          />

          <TextField
            label="自己紹介"
            fullWidth
            multiline
            rows={4}
            value={formData.bio}
            onChange={(e) => handleChange("bio", e.target.value)}
            helperText="簡単な自己紹介を入力してください（任意）"
            disabled={loading}
          />

          <TextField
            label="ウェブサイトURL"
            fullWidth
            type="url"
            value={formData.websiteUrl}
            onChange={(e) => handleChange("websiteUrl", e.target.value)}
            helperText="個人サイトやSNSのURLを入力してください（任意）"
            disabled={loading}
            placeholder="https://example.com"
          />

          <TextField
            label="Twitterハンドル"
            fullWidth
            value={formData.twitterHandle}
            onChange={(e) => handleChange("twitterHandle", e.target.value)}
            helperText="@なしのTwitterユーザー名を入力してください（任意）"
            disabled={loading}
            placeholder="username"
          />

          {profile?.birthdate && (
            <Box>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                誕生日:{" "}
                {new Date(profile.birthdate).toLocaleDateString("ja-JP", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </Typography>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.birthdateVisible || false}
                    onChange={(e) =>
                      handleChange("birthdateVisible", e.target.checked)
                    }
                    disabled={loading}
                  />
                }
                label="誕生日を公開する"
              />
              <Typography
                variant="caption"
                color="text.secondary"
                display="block"
              >
                誕生日の公開/非公開を選択できます
              </Typography>
            </Box>
          )}

          <Stack direction="row" spacing={2} justifyContent="flex-end">
            {onCancel && (
              <Button
                variant="outlined"
                startIcon={<CancelIcon />}
                onClick={onCancel}
                disabled={loading}
              >
                キャンセル
              </Button>
            )}

            <Button
              type="submit"
              variant="contained"
              startIcon={
                loading ? <CircularProgress size={20} /> : <SaveIcon />
              }
              disabled={loading}
            >
              {loading ? "保存中..." : "保存"}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Card>
  );
}
