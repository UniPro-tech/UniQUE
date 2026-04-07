"use client";

import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Snackbar,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { bulkAssignUsersToRole } from "./action";

interface Props {
  roleId: string;
}

export default function RoleBulkAssignButton({ roleId }: Props) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [snack, setSnack] = useState<{
    msg: string;
    severity: "success" | "error";
  } | null>(null);
  const router = useRouter();

  const handleConfirm = async () => {
    setOpen(false);
    setLoading(true);
    try {
      const res = await bulkAssignUsersToRole({ roleId });
      if (res.error) {
        const msg = `一括付与に失敗しました: ${res.error} `;
        setSnack({
          msg,
          severity: "error",
        });
      } else {
        setSnack({
          msg: `一括付与が完了しました。`,
          severity: "success",
        });
        router.refresh();
      }
    } catch (err) {
      console.error(err);
      setSnack({ msg: "一括付与中にエラーが発生しました", severity: "error" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <Button
        variant="outlined"
        onClick={() => setOpen(true)}
        disabled={loading}
      >
        {loading ? <CircularProgress size={18} /> : "全てのユーザーに一括付与"}
      </Button>

      <Dialog open={open} onClose={() => setOpen(false)}>
        <DialogTitle>全ユーザーに一括付与</DialogTitle>
        <DialogContent>
          <DialogContentText>
            このロールをステータスが active または established
            の既存ユーザー全員に付与しますか？ この操作は取り消せません。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)} disabled={loading}>
            キャンセル
          </Button>
          <Button
            variant="contained"
            onClick={handleConfirm}
            disabled={loading}
          >
            実行
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={!!snack}
        autoHideDuration={6000}
        onClose={(_event, reason) => {
          if (reason === "clickaway") return;
          setSnack(null);
        }}
      >
        {snack ? (
          <Alert
            severity={snack.severity}
            onClose={() => setSnack(null)}
            sx={{ width: "100%" }}
            variant="filled"
          >
            {snack.msg}
          </Alert>
        ) : undefined}
      </Snackbar>
    </>
  );
}
