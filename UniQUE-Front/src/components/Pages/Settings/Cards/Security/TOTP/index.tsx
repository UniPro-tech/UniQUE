"use client";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { enqueueSnackbar } from "notistack";
import { QRCodeCanvas } from "qrcode.react";
import { useActionState, useEffect, useState } from "react";
import type { UserData } from "@/classes/types/User";
import { disableTOTP, generateTOTP, verifyTOTP } from "./action";

export default function TOTPSection({
  user,
  csrfToken,
}: {
  user: UserData;
  csrfToken: string;
}) {
  const isEnabled = user.isTotpEnabled === true;

  const [genResult, genAction] = useActionState(
    generateTOTP,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    null as null | any,
  );
  const [verifyResult, verifyAction] = useActionState(
    verifyTOTP,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    null as null | { valid: boolean } | any,
  );
  const [disableResult, disableAction] = useActionState(
    disableTOTP,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    null as null | any,
  );

  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [openDialog, setOpenDialog] = useState(false);
  const [openDisableDialog, setOpenDisableDialog] = useState(false);
  const [disablePassword, setDisablePassword] = useState("");
  const [localEnabled, setLocalEnabled] = useState(isEnabled);
  const [isFinished, setIsFinished] = useState(false);

  useEffect(() => {
    if (genResult?.error) {
      enqueueSnackbar(genResult.error, { variant: "error" });
    } else if (genResult?.secret) {
      Promise.resolve().then(() => {
        setIsFinished(true);
      });
      enqueueSnackbar(
        "TOTP シークレットを生成しました。コードを入力して有効化してください。",
        { variant: "success" },
      );
      // update state in microtask to avoid cascading renders warning
      Promise.resolve().then(() => {
        setOpenDialog(false);
        setPassword("");
      });
    }
  }, [genResult]);

  useEffect(() => {
    if (verifyResult?.error) {
      enqueueSnackbar(verifyResult.error, { variant: "error" });
    } else if (verifyResult?.valid) {
      enqueueSnackbar("二段階認証を有効化しました。", { variant: "success" });
      Promise.resolve().then(() => {
        setLocalEnabled(true);
        setIsFinished(false);
      });
    } else if (verifyResult && verifyResult.valid === false) {
      enqueueSnackbar("コードが無効です。再度ご確認ください。", {
        variant: "error",
      });
    }
  }, [verifyResult]);

  useEffect(() => {
    if (disableResult?.error) {
      enqueueSnackbar(disableResult.error, { variant: "error" });
    } else if (disableResult?.success) {
      enqueueSnackbar(disableResult.message || "二段階認証を無効化しました。", {
        variant: "success",
      });
      Promise.resolve().then(() => {
        setOpenDisableDialog(false);
        setDisablePassword("");
        setLocalEnabled(false);
      });
    }
  }, [disableResult]);

  const uri = genResult?.uri || null;

  return (
    <Stack spacing={2} sx={{ p: 2 }}>
      <Stack direction="row" alignItems="center" spacing={2}>
        <Typography variant="h6" component={"h4"}>
          二段階認証（TOTP）
        </Typography>
        <Divider sx={{ flexGrow: 1 }} />
      </Stack>

      <Box>
        <Typography variant="body2" color="text.secondary">
          現在の状態: {localEnabled ? "有効" : "無効"}
        </Typography>
      </Box>

      {!localEnabled && !genResult?.secret && (
        <>
          <Button
            variant="outlined"
            sx={{ alignSelf: "flex-start" }}
            onClick={() => setOpenDialog(true)}
          >
            セットアップを開始
          </Button>

          <Dialog
            open={openDialog}
            onClose={() => setOpenDialog(false)}
            fullWidth
            maxWidth="xs"
          >
            <DialogTitle>二段階認証のセットアップ</DialogTitle>
            <DialogContent>
              <Box component="form" action={genAction} id="totp-gen-form">
                <input type="hidden" name="csrfToken" value={csrfToken} />
                <input type="hidden" name="user_id" value={user.id} />
                <TextField
                  autoFocus
                  margin="dense"
                  label="パスワード"
                  name="password"
                  type="password"
                  fullWidth
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </Box>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setOpenDialog(false)}>キャンセル</Button>
              <Button type="submit" form="totp-gen-form" variant="contained">
                確認して生成
              </Button>
            </DialogActions>
          </Dialog>
        </>
      )}

      {localEnabled && (
        <>
          <Button
            variant="outlined"
            color="error"
            sx={{ alignSelf: "flex-start" }}
            onClick={() => setOpenDisableDialog(true)}
          >
            無効化
          </Button>

          <Dialog
            open={openDisableDialog}
            onClose={() => setOpenDisableDialog(false)}
            fullWidth
            maxWidth="xs"
          >
            <DialogTitle>二段階認証の無効化</DialogTitle>
            <DialogContent>
              <Box
                component="form"
                action={disableAction}
                id="totp-disable-form"
              >
                <input type="hidden" name="csrfToken" value={csrfToken} />
                <TextField
                  autoFocus
                  margin="dense"
                  label="パスワード"
                  name="password"
                  type="password"
                  fullWidth
                  value={disablePassword}
                  onChange={(e) => setDisablePassword(e.target.value)}
                />
              </Box>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setOpenDisableDialog(false)}>
                キャンセル
              </Button>
              <Button
                type="submit"
                form="totp-disable-form"
                variant="contained"
              >
                無効化する
              </Button>
            </DialogActions>
          </Dialog>
        </>
      )}

      {genResult?.secret && isFinished && (
        <Stack spacing={2}>
          <Box sx={{ display: "flex", gap: 2, alignItems: "center" }}>
            {uri ? <QRCodeCanvas value={uri} size={200} /> : null}
            <Box>
              <Typography variant="body2">
                シークレット: {genResult.secret}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                QR またはシークレットを認証アプリに登録してください。
              </Typography>
            </Box>
          </Box>

          <Box component="form" action={verifyAction}>
            <input type="hidden" name="csrfToken" value={csrfToken} />
            <TextField
              label="認証コード"
              name="code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              fullWidth
            />
            <Button type="submit" variant="contained" sx={{ mt: 1 }}>
              有効化
            </Button>
          </Box>
        </Stack>
      )}
    </Stack>
  );
}
