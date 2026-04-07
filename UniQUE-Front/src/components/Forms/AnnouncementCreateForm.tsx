"use client";

import { Cancel as CancelIcon, Save as SaveIcon } from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  CircularProgress,
  FormControlLabel,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { createAnnouncement } from "@/app/dashboard/announcements/new/action";

interface AnnouncementCreateFormProps {
  onCancel?: () => void;
}

interface CreateAnnouncementFormData {
  title: string;
  content: string;
  isPinned?: boolean;
}

export default function AnnouncementCreateForm({
  onCancel,
}: AnnouncementCreateFormProps) {
  const router = useRouter();

  const [formData, setFormData] = useState<CreateAnnouncementFormData>({
    title: "",
    content: "",
    isPinned: false,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await createAnnouncement({
        title: formData.title,
        content: formData.content,
        isPinned: formData.isPinned,
      });
      if (res.success) {
        router.push("/dashboard/announcements");
      } else {
        setError(res.error || "作成に失敗しました");
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg || "作成に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    if (onCancel) onCancel();
    else router.push("/dashboard/announcements");
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        お知らせ新規作成
      </Typography>

      <Box component="form" onSubmit={handleSubmit} noValidate>
        <Stack spacing={3}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          <TextField
            label="タイトル"
            fullWidth
            required
            value={formData.title}
            onChange={(e) =>
              setFormData((p) => ({ ...p, title: e.target.value }))
            }
            disabled={loading}
          />

          <TextField
            label="本文"
            fullWidth
            required
            multiline
            rows={6}
            value={formData.content}
            onChange={(e) =>
              setFormData((p) => ({ ...p, content: e.target.value }))
            }
            disabled={loading}
          />

          <FormControlLabel
            control={
              <Checkbox
                checked={!!formData.isPinned}
                onChange={(e) =>
                  setFormData((p) => ({ ...p, isPinned: e.target.checked }))
                }
                disabled={loading}
              />
            }
            label="ピン留めする"
          />

          <Stack direction="row" spacing={2} justifyContent="flex-end">
            <Button
              variant="outlined"
              startIcon={<CancelIcon />}
              onClick={handleCancel}
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
