"use client";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { resetPassword } from "./action";

export default function PasswordResetAdmin({
  open,
  onClose,
  userId,
}: {
  open: boolean;
  onClose: () => void;
  userId: string;
}) {
  const { enqueueSnackbar } = useSnackbar();
  return (
    <Dialog open={open} onClose={onClose}>
      <form
        action={async (formdata: FormData) => {
          const newPassword = formdata.get("newPassword");
          if (!newPassword || typeof newPassword !== "string") return;
          try {
            await resetPassword(userId, newPassword);
            enqueueSnackbar("パスワードリセットを行いました。", {
              variant: "success",
            });
            onClose();
          } catch {
            enqueueSnackbar("パスワードリセットに失敗しました。", {
              variant: "error",
            });
          }
        }}
      >
        <DialogContent
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: 2,
            width: "100%",
          }}
        >
          <DialogTitle>管理者によるパスワードリセット</DialogTitle>
          <DialogContentText>
            パスワードをリセットするには、新しいパスワードを入力してください。
          </DialogContentText>
          <TextField
            label={"新しいパスワード"}
            name={"newPassword"}
            type="password"
            autoComplete="new-password"
            required
          />
        </DialogContent>
        <DialogActions sx={{ pb: 3, px: 3 }}>
          <Button
            onClick={() => {
              onClose();
            }}
          >
            キャンセル
          </Button>
          <Button variant="contained" type="submit" color="error">
            リセット
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
