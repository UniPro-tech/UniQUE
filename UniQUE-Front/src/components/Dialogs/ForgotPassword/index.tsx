"use client";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import OutlinedInput from "@mui/material/OutlinedInput";
import { SnackbarProvider, useSnackbar } from "notistack";
import * as React from "react";
import { submitForgotPassword } from "./action";

interface ForgotPasswordProps {
  open: boolean;
  handleClose: () => void;
  csrfToken: string;
}

export default function ForgotPassword({
  open,
  handleClose,
  csrfToken,
}: ForgotPasswordProps) {
  const [email, setEmail] = React.useState("");
  const { enqueueSnackbar } = useSnackbar();
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Dialog
        open={open}
        onClose={handleClose}
        onSubmit={async (formevent) => {
          formevent.preventDefault();
          const res = await submitForgotPassword(email);
          enqueueSnackbar(res.message, { variant: res.status });
          if (res.status === "success") {
            handleClose();
          }
        }}
        slotProps={{
          paper: {
            sx: { backgroundImage: "none" },
          },
        }}
      >
        <DialogTitle>パスワードリセット</DialogTitle>
        <DialogContent
          sx={{
            display: "flex",
            flexDirection: "column",
            gap: 2,
            width: "100%",
          }}
        >
          <DialogContentText>
            アカウントに登録されている外部メールアドレスを入力してください。パスワードをリセットするためのリンクをお送りします。
          </DialogContentText>
          <input type="hidden" name="csrfToken" value={csrfToken} />
          <OutlinedInput
            autoFocus
            required
            margin="dense"
            id="email"
            name="email"
            label="メールアドレス"
            placeholder="example@example.com"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            fullWidth
          />
        </DialogContent>
        <DialogActions sx={{ pb: 3, px: 3 }}>
          <Button onClick={handleClose}>キャンセル</Button>
          <Button variant="contained" type="submit">
            送信
          </Button>
        </DialogActions>
      </Dialog>
    </SnackbarProvider>
  );
}
