"use client";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Typography,
} from "@mui/material";
import { useActionState, useEffect } from "react";
import type { FormStatus } from "@/components/Pages/Settings/Cards/Base";
import { updateBirthdate } from "./action";

export default function BirthdateGuard({
  mustSetBirthdate,
  csrfToken,
  initialBirthdate,
}: {
  mustSetBirthdate: boolean;
  csrfToken: string;
  initialBirthdate: string;
}) {
  const [state, action, isPending] = useActionState(updateBirthdate, {
    birthdate: initialBirthdate,
    status: null as FormStatus | null,
  });

  useEffect(() => {
    if (state.status?.status === "success") {
      // nothing to do: dialog will close automatically
    }
  }, [state.status]);

  if (!mustSetBirthdate) return null;

  const open = !state.birthdate;

  return (
    <Dialog
      open={open}
      onClose={() => undefined}
      disableEscapeKeyDown
      PaperProps={{ component: "form", action }}
    >
      <DialogTitle>生年月日の設定が必要です</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
        <Typography variant="body2">
          未成年者保護のため、生年月日の登録をお願いします。
        </Typography>
        <input type="hidden" name="csrfToken" value={csrfToken} />
        <TextField
          required
          label="生年月日"
          type="date"
          name="birthdate"
          defaultValue={state.birthdate}
          InputLabelProps={{ shrink: true }}
          disabled={isPending}
        />
        {state.status && (
          <Typography
            variant="body2"
            color={state.status.status === "error" ? "error" : "success"}
          >
            {state.status.message}
          </Typography>
        )}
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 3 }}>
        <Button type="submit" variant="contained" disabled={isPending}>
          設定する
        </Button>
      </DialogActions>
    </Dialog>
  );
}
