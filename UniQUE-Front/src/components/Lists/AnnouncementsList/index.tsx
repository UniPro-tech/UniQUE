"use client";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import PushPinIcon from "@mui/icons-material/PushPin";
import VisibilityIcon from "@mui/icons-material/Visibility";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Grid,
  Snackbar,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import Link from "next/link";
import React from "react";
import type { AnnouncementData } from "@/classes/Announcement";
import { deleteAnnouncement } from "./actions/delete";
import { pinAnnouncement } from "./actions/pin";

type Props = {
  initial?: AnnouncementData[];
  canUpdate?: boolean;
  canPin?: boolean;
  canDelete?: boolean;
  isAdmin?: boolean;
};

export default function AnnouncementsList({
  initial,
  canUpdate = false,
  canPin = false,
  canDelete = false,
  isAdmin = false,
}: Props) {
  const [items, setItems] = React.useState<AnnouncementData[]>(
    () => initial || [],
  );
  const [busy, setBusy] = React.useState<Record<string, boolean>>({});

  const [editingId, setEditingId] = React.useState<string | null>(null);
  const [editValues, setEditValues] = React.useState<{
    title: string;
    content: string;
  }>({ title: "", content: "" });
  const [editDialogOpen, setEditDialogOpen] = React.useState(false);

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [deleteTarget, setDeleteTarget] = React.useState<string | null>(null);

  const [snackbar, setSnackbar] = React.useState<{
    open: boolean;
    message: string;
    severity: "success" | "error" | "info" | "warning";
  }>({ open: false, message: "", severity: "info" });

  const cancelEdit = () => {
    setEditingId(null);
    setEditValues({ title: "", content: "" });
    setEditDialogOpen(false);
  };

  const saveEdit = async (id: string) => {
    setBusy((s) => ({ ...s, [id]: true }));
    try {
      const res = await fetch(`/api/announcements/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          title: editValues.title,
          content: editValues.content,
        }),
      });
      if (!res.ok) throw new Error("failed to update");
      setItems((prev) =>
        prev.map((p) =>
          p.id === id
            ? { ...p, title: editValues.title, content: editValues.content }
            : p,
        ),
      );
      setEditingId(null);
      setEditDialogOpen(false);
      setSnackbar({ open: true, message: "更新しました", severity: "success" });
    } catch (err) {
      console.error(err);
      setSnackbar({
        open: true,
        message: "更新に失敗しました",
        severity: "error",
      });
    } finally {
      setBusy((s) => ({ ...s, [id]: false }));
    }
  };

  const confirmDelete = (id: string) => {
    setDeleteTarget(id);
    setDeleteDialogOpen(true);
  };

  const performDelete = async (id: string) => {
    setDeleteDialogOpen(false);
    if (!id) return;
    setBusy((s) => ({ ...s, [id]: true }));
    try {
      await deleteAnnouncement(id);
      setItems((prev) => prev.filter((p) => p.id !== id));
      setSnackbar({ open: true, message: "削除しました", severity: "success" });
    } catch (err) {
      console.error(err);
      setSnackbar({
        open: true,
        message: "削除に失敗しました",
        severity: "error",
      });
    } finally {
      setBusy((s) => ({ ...s, [id]: false }));
      setDeleteTarget(null);
    }
  };

  const handlePin = async (id: string, pin: boolean) => {
    setBusy((s) => ({ ...s, [id]: true }));
    try {
      await pinAnnouncement(id, pin);
      setItems((prev) =>
        prev.map((p) => (p.id === id ? { ...p, isPinned: pin } : p)),
      );
      setSnackbar({
        open: true,
        message: pin ? "ピンしました" : "ピンを外しました",
        severity: "success",
      });
    } catch (err) {
      console.error(err);
      setSnackbar({
        open: true,
        message: "操作に失敗しました",
        severity: "error",
      });
    } finally {
      setBusy((s) => ({ ...s, [id]: false }));
    }
  };

  if (!items || items.length === 0) {
    return (
      <Box textAlign="center" py={6}>
        <Typography variant="h6">お知らせはありません</Typography>
        <Typography variant="body2" color="text.secondary">
          新しいお知らせが追加されるとここに表示されます。
        </Typography>
      </Box>
    );
  }

  const effectiveCanUpdate = !!isAdmin || canUpdate;
  const effectiveCanPin = !!isAdmin || canPin;
  const effectiveCanDelete = !!isAdmin || canDelete;

  return (
    <>
      <Grid
        container
        spacing={2}
        justifyContent="start"
        alignItems="start"
        id="announcements-grid"
        columns={12}
      >
        {items.map((a) => {
          const preview =
            a.content && a.content.length > 200
              ? `${a.content.slice(0, 200)}…`
              : a.content;

          return (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={a.id}>
              <Card variant="outlined">
                <CardHeader
                  avatar={
                    <Avatar
                      aria-label={
                        a.createdBy?.profile?.displayName ||
                        a.createdBy?.customId ||
                        a.createdBy?.id ||
                        "system"
                      }
                    >
                      {a.createdBy?.profile?.displayName?.[0] ||
                        a.createdBy?.customId?.[0] ||
                        a.createdBy?.id?.[0] ||
                        "S"}
                    </Avatar>
                  }
                  title={
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Typography variant="h6" component="div">
                        {a.title}
                      </Typography>
                      {a.isPinned && (
                        <Chip label="ピン" size="small" color="primary" />
                      )}
                    </Stack>
                  }
                  subheader={`作成者: ${
                    a.createdBy?.profile?.displayName ||
                    a.createdBy?.customId ||
                    a.createdBy?.id ||
                    "system"
                  } ・ ${new Date(a.createdAt).toLocaleString()}`}
                />
                <CardContent>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    component="div"
                    sx={{ whiteSpace: "pre-wrap" }}
                  >
                    {preview}
                  </Typography>
                </CardContent>
                <Divider />
                <CardActions>
                  <Link
                    href={`/dashboard/announcements/${a.id}`}
                    style={{ textDecoration: "none" }}
                  >
                    <Button
                      startIcon={<VisibilityIcon />}
                      size="small"
                      disabled={!!busy[a.id]}
                    >
                      詳細
                    </Button>
                  </Link>
                  {effectiveCanUpdate && (
                    <Link
                      href={`/dashboard/announcements/${a.id}/edit`}
                      style={{ textDecoration: "none" }}
                    >
                      <Button
                        startIcon={<EditIcon />}
                        size="small"
                        disabled={!!busy[a.id]}
                      >
                        編集
                      </Button>
                    </Link>
                  )}
                  {effectiveCanDelete && (
                    <Button
                      startIcon={<DeleteIcon />}
                      size="small"
                      color="error"
                      onClick={() => confirmDelete(a.id)}
                      disabled={!!busy[a.id]}
                    >
                      削除
                    </Button>
                  )}
                  {effectiveCanPin && (
                    <Button
                      startIcon={<PushPinIcon />}
                      size="small"
                      variant={a.isPinned ? "contained" : "outlined"}
                      onClick={() => handlePin(a.id, !a.isPinned)}
                      disabled={!!busy[a.id]}
                    >
                      {a.isPinned ? "ピン外し" : "ピン"}
                    </Button>
                  )}
                </CardActions>
              </Card>
            </Grid>
          );
        })}
      </Grid>

      <Dialog
        open={editDialogOpen}
        onClose={cancelEdit}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle>お知らせを編集</DialogTitle>
        <DialogContent>
          <TextField
            label="タイトル"
            fullWidth
            margin="normal"
            value={editValues.title}
            onChange={(e) =>
              setEditValues((v) => ({ ...v, title: e.target.value }))
            }
          />
          <TextField
            label="内容"
            fullWidth
            margin="normal"
            multiline
            rows={6}
            value={editValues.content}
            onChange={(e) =>
              setEditValues((v) => ({ ...v, content: e.target.value }))
            }
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={cancelEdit}>キャンセル</Button>
          <Button
            variant="contained"
            onClick={() => editingId && saveEdit(editingId)}
            disabled={editingId ? !!busy[editingId] : true}
          >
            保存
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
      >
        <DialogTitle>確認</DialogTitle>
        <DialogContent>本当にお知らせを削除しますか？</DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>キャンセル</Button>
          <Button
            variant="contained"
            color="error"
            onClick={() => deleteTarget && performDelete(deleteTarget)}
          >
            削除
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar((s) => ({ ...s, open: false }))}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar((s) => ({ ...s, open: false }))}
          variant="filled"
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </>
  );
}
