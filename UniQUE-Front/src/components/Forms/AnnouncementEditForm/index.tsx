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
  Snackbar,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import type React from "react";
import { useState } from "react";
import { updateAnnouncement } from "./action";

type Props = {
  id: string;
  initial: { title: string; content: string; isPinned?: boolean };
};

export default function AnnouncementEditForm({ id, initial }: Props) {
  const router = useRouter();
  const [formData, setFormData] = useState({
    title: initial.title,
    content: initial.content,
    isPinned: !!initial.isPinned,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [snackOpen, setSnackOpen] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await updateAnnouncement({
        id,
        title: formData.title,
        content: formData.content,
        isPinned: formData.isPinned,
      });
      setSnackOpen(true);
      setTimeout(() => router.push(`/dashboard/announcements/${id}`), 800);
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      setError(msg || "更新に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        お知らせを編集
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
                checked={formData.isPinned}
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
              onClick={() => router.back()}
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
              {loading ? "保存中..." : "保存"}
            </Button>
          </Stack>
        </Stack>
      </Box>

      <Snackbar
        open={snackOpen}
        autoHideDuration={3000}
        onClose={() => setSnackOpen(false)}
      >
        <Alert
          severity="success"
          variant="filled"
          onClose={() => setSnackOpen(false)}
        >
          更新しました
        </Alert>
      </Snackbar>
    </Card>
  );
}
