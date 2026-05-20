import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import { enqueueSnackbar, SnackbarProvider } from "notistack";
import * as React from "react";
import type { UserData } from "@/classes/types/User";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { rejectRegistApplyAction } from "./actions/rejectRegistApplyAction";

interface RejectRegistApplyProps {
  open: boolean;
  handleClose: () => void;
  user: UserData | null;
  deleteRowAction: (userId: string) => void;
}

export default function RejectDialog({
  open,
  handleClose,
  user,
  deleteRowAction,
}: RejectRegistApplyProps) {
  const [state, action, isPending] = React.useActionState(
    rejectRegistApplyAction,
    null as null | FormStatus,
  );
  React.useEffect(() => {
    if (state) {
      enqueueSnackbar(state.message, { variant: state.status });
      if (state.status === "success") {
        // biome-ignore lint/style/noNonNullAssertion: 上流で保証されている
        deleteRowAction(user!.id);
        handleClose();
      }
    }
  }, [handleClose, state, deleteRowAction, user?.id]);

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
        <form action={action} id="reject-regist-apply-data-dialog">
          <DialogTitle>メンバーの却下</DialogTitle>
          <DialogContent
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
              width: "100%",
            }}
          >
            <DialogContentText>
              本当にこのユーザーの申請を却下しますか？この操作は取り消せません。
            </DialogContentText>
            <DialogContentText>
              カスタムID:{" "}
              {user?.customId || user?.profile?.displayName || "不明なユーザー"}
            </DialogContentText>
            {/** biome-ignore lint/style/noNonNullAssertion: 上流で保証されている */}
            <input type="hidden" name="userId" value={user!.id} />
          </DialogContent>
          <DialogActions sx={{ pb: 3, px: 3 }}>
            <Button onClick={handleClose} disabled={isPending}>
              キャンセル
            </Button>
            <Button variant="contained" type="submit" disabled={isPending}>
              却下
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </SnackbarProvider>
  );
}
