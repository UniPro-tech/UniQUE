"use client";

import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import OutlinedInput from "@mui/material/OutlinedInput";
import { useSnackbar } from "notistack";
import { useState } from "react";
import { submitForgotPassword } from "./action";

interface ForgotPasswordProps {
  open: boolean;
  handleClose: () => void;
}

export default function ForgotPassword({
  open,
  handleClose,
}: ForgotPasswordProps) {
  const [email, setEmail] = useState("");
  const { enqueueSnackbar } = useSnackbar();
  return (
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
          component: "form",
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
          登録されている外部メールアドレスを入力してください。パスワードリセットの手順を記載したメールが送信されます。
        </DialogContentText>
        <OutlinedInput
          autoFocus
          required
          margin="dense"
          id="email"
          name="email"
          label="Email address"
          placeholder="Email address"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          fullWidth
        />
      </DialogContent>
      <DialogActions sx={{ pb: 3, px: 3 }}>
        <Button onClick={handleClose}>Cancel</Button>
        <Button variant="contained" type="submit">
          続行
        </Button>
      </DialogActions>
    </Dialog>
  );
}
