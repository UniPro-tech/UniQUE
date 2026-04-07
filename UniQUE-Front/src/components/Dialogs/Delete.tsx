"use client";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import { enqueueSnackbar, SnackbarProvider } from "notistack";
import * as React from "react";
import { useEffect } from "react";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";

interface DeleteDialogProps {
  open: boolean;
  dataAction: (
    _prevState: FormStatus | null,
    formData: FormData | null,
  ) => FormStatus | null | Promise<FormStatus | null>;
  handleClose: () => void;
  title?: string;
}

export default function DeleteDialog({
  open,
  dataAction,
  handleClose,
  title,
}: DeleteDialogProps) {
  const [state, action] = React.useActionState(
    dataAction,
    null as null | FormStatus,
  );
  useEffect(() => {
    if (state) {
      enqueueSnackbar(state.message, { variant: state.status });
      handleClose();
    }
  }, [state, handleClose]);
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <Dialog
        open={open}
        onClose={handleClose}
        slotProps={{
          paper: {
            sx: { backgroundImage: "none" },
          },
        }}
      >
        <form action={action} id="delete-data-dialog">
          <DialogTitle>{title || "データ"}を削除しますか？</DialogTitle>
          <DialogContent
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
              width: "100%",
            }}
          >
            <DialogContentText>
              本当にデータを削除してもよろしいですか？
              この操作は元に戻せません。
            </DialogContentText>
          </DialogContent>
          <DialogActions sx={{ pb: 3, px: 3 }}>
            <Button onClick={handleClose}>キャンセル</Button>
            <Button variant="contained" type="submit" color="error">
              削除
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </SnackbarProvider>
  );
}
