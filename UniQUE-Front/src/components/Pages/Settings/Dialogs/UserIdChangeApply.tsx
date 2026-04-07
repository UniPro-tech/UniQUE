"use client";

import { TextField } from "@mui/material";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import { SnackbarProvider, useSnackbar } from "notistack";
import * as React from "react";
import type { UserData } from "@/classes/types/User";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { userIdChangeApplyAction } from "./actions/approveRegistApplyAction";

interface UserIdChangeApplyProps {
  open: boolean;
  handleClose: () => void;
  user: UserData | null;
}

export default function UserIdChangeApply({
  open,
  handleClose,
  user,
}: UserIdChangeApplyProps) {
  const [state, action, isPending] = React.useActionState(
    userIdChangeApplyAction,
    null as null | FormStatus,
  );
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <InnerDialog
        open={open}
        handleClose={handleClose}
        user={user}
        state={state}
        action={action}
        isPending={isPending}
      />
    </SnackbarProvider>
  );

  function InnerDialog({
    open,
    handleClose,
    user,
    state,
    action,
    isPending,
  }: UserIdChangeApplyProps & {
    state: FormStatus | null;
    action: (formData: FormData) => void | Promise<void>;
    isPending: boolean;
  }) {
    const { enqueueSnackbar } = useSnackbar();

    React.useEffect(() => {
      if (state) {
        enqueueSnackbar(state.message, { variant: state.status });
        handleClose();
      }
    }, [state, handleClose, enqueueSnackbar]);

    return (
      <Dialog
        open={open}
        onClose={handleClose}
        slotProps={{
          paper: {
            sx: { backgroundImage: "none" },
          },
        }}
      >
        <form action={action} id="approve-regist-apply-data-dialog">
          <DialogTitle>カスタムIDの変更申請</DialogTitle>
          <DialogContent
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
              width: "100%",
            }}
          >
            <DialogContentText>
              <strong>カスタムIDの変更は原則としてできません。</strong>
              <br />
              特別な事情がある場合に限り、管理者が承認した場合のみ変更されます。
              <br />
              また、原則として、<strong>メールアドレスも変更となります</strong>
              ので、ご了承ください。
            </DialogContentText>
            <input type="hidden" name="userId" value={user?.id || ""} />
            <TextField
              label="新しいカスタムID"
              type="text"
              name="newCustomId"
              required
              fullWidth
              variant="outlined"
              placeholder="hogehoge"
              disabled={isPending}
              helperText="@を除いたカスタムIDを入力してください。"
            />
            <TextField
              label="変更理由"
              type="text"
              name="changeReason"
              required
              fullWidth
              variant="outlined"
              placeholder="変更理由を入力してください"
              disabled={isPending}
              helperText="変更理由を入力してください。"
            />
          </DialogContent>
          <DialogActions sx={{ pb: 3, px: 3 }}>
            <Button onClick={handleClose} disabled={isPending}>
              キャンセル
            </Button>
            <Button variant="contained" type="submit" disabled={isPending}>
              申請を送信
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    );
  }
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
        <form action={action} id="approve-regist-apply-data-dialog">
          <DialogTitle>カスタムIDの変更申請</DialogTitle>
          <DialogContent
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
              width: "100%",
            }}
          >
            <DialogContentText>
              <strong>カスタムIDの変更は原則としてできません。</strong>
              <br />
              特別な事情がある場合に限り、管理者が承認した場合のみ変更されます。
              <br />
              また、原則として、<strong>メールアドレスも変更となります</strong>
              ので、ご了承ください。
            </DialogContentText>
            <input type="hidden" name="userId" value={user?.id || ""} />
            <TextField
              label="新しいカスタムID"
              type="text"
              name="newCustomId"
              required
              fullWidth
              variant="outlined"
              placeholder="hogehoge"
              disabled={isPending}
              helperText="@を除いたカスタムIDを入力してください。"
            />
            <TextField
              label="変更理由"
              type="text"
              name="changeReason"
              required
              fullWidth
              variant="outlined"
              placeholder="変更理由を入力してください"
              disabled={isPending}
              helperText="変更理由を入力してください。"
            />
          </DialogContent>
          <DialogActions sx={{ pb: 3, px: 3 }}>
            <Button onClick={handleClose} disabled={isPending}>
              キャンセル
            </Button>
            <Button variant="contained" type="submit" disabled={isPending}>
              申請を送信
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </SnackbarProvider>
  );
}
