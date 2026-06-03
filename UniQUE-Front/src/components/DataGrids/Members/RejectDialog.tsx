import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import { enqueueSnackbar } from "notistack";
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

  const submittingUserIdRef = React.useRef<string | null>(null);
  const handleSubmit = React.useCallback(() => {
    if (user) {
      submittingUserIdRef.current = user.id;
    }
  }, [user]);

  React.useEffect(() => {
    if (state) {
      enqueueSnackbar(state.message, { variant: state.status });
      if (state.status === "success") {
        const userId = submittingUserIdRef.current;
        if (!userId) return;
        deleteRowAction(userId);
        submittingUserIdRef.current = null; // リセットして再実行を防ぐ
        handleClose();
      }
    }
  }, [handleClose, state, deleteRowAction]);

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
      <form
        action={action}
        id="reject-regist-apply-data-dialog"
        onSubmit={handleSubmit}
      >
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
          <input type="hidden" name="userId" value={user?.id} />
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
  );
}
