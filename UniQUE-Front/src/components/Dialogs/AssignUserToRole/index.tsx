"use client";
import {
  Alert,
  Autocomplete,
  Box,
  CircularProgress,
  TextField,
} from "@mui/material";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import { enqueueSnackbar, SnackbarProvider } from "notistack";
import * as React from "react";
import type { UserData } from "@/classes/types/User";
import {
  assignUserToRole,
  getAllUsers,
} from "@/components/Dialogs/AssignUserToRole/action";

interface AssignUserToRoleDialogProps {
  open: boolean;
  roleId: string;
  roleName: string;
  currentUserIds: string[];
  handleClose: () => void;
  onSuccess?: () => void;
}

export default function AssignUserToRoleDialog({
  open,
  roleId,
  roleName,
  currentUserIds,
  handleClose,
  onSuccess,
}: AssignUserToRoleDialogProps) {
  const [users, setUsers] = React.useState<UserData[]>([]);
  const [selectedUser, setSelectedUser] = React.useState<UserData | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  // ユーザー一覧を取得
  React.useEffect(() => {
    if (open) {
      setLoading(true);
      setError(null);
      setSelectedUser(null);
      getAllUsers()
        .then((allUsers) => {
          // 既に割り当てられているユーザーを除外
          const availableUsers = allUsers.filter(
            (u) => !currentUserIds.includes(u.id),
          );
          setUsers(availableUsers);
        })
        .catch(() => {
          setError("ユーザー一覧の取得に失敗しました");
        })
        .finally(() => {
          setLoading(false);
        });
    }
  }, [open, currentUserIds]);

  const handleSubmit = async () => {
    if (!selectedUser) {
      setError("ユーザーを選択してください");
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const result = await assignUserToRole(roleId, selectedUser.id);

      if (result.success) {
        enqueueSnackbar(
          `${selectedUser.profile?.displayName || selectedUser.customId}を${roleName}に追加しました`,
          { variant: "success" },
        );
        handleClose();
        if (onSuccess) {
          onSuccess();
        }
      } else {
        setError(result.error || "割り当てに失敗しました");
      }
    } catch (err) {
      setError("予期しないエラーが発生しました");
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const getUserLabel = (user: UserData) => {
    const displayName = user.profile?.displayName || user.customId;
    return `${displayName} (${user.email})`;
  };

  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Dialog
        open={open}
        onClose={handleClose}
        maxWidth="sm"
        fullWidth
        slotProps={{
          paper: {
            sx: { backgroundImage: "none" },
          },
        }}
      >
        <DialogTitle>ユーザーを追加</DialogTitle>
        <DialogContent
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: 2,
            width: "100%",
            minHeight: 200,
          }}
        >
          <DialogContentText>
            「{roleName}」ロールに追加するユーザーを選択してください
          </DialogContentText>

          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          {loading ? (
            <Box
              sx={{
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                py: 4,
              }}
            >
              <CircularProgress />
            </Box>
          ) : users.length === 0 ? (
            <Alert severity="info">
              割り当て可能なユーザーがいません。すべてのユーザーがこのロールに既に割り当てられています。
            </Alert>
          ) : (
            <Autocomplete
              options={users}
              getOptionLabel={getUserLabel}
              value={selectedUser}
              onChange={(_event, newValue) => {
                setSelectedUser(newValue);
                setError(null);
              }}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label="ユーザーを検索"
                  placeholder="名前またはメールアドレスで検索"
                  disabled={submitting}
                />
              )}
              noOptionsText="該当するユーザーが見つかりません"
              disabled={submitting}
            />
          )}
        </DialogContent>
        <DialogActions sx={{ pb: 3, px: 3 }}>
          <Button onClick={handleClose} disabled={submitting}>
            キャンセル
          </Button>
          <Button
            variant="contained"
            onClick={handleSubmit}
            disabled={!selectedUser || submitting || users.length === 0}
            startIcon={submitting ? <CircularProgress size={20} /> : null}
          >
            {submitting ? "追加中..." : "追加"}
          </Button>
        </DialogActions>
      </Dialog>
    </SnackbarProvider>
  );
}
