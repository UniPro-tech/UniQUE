"use client";
import {
  GppBad as GppBadIcon,
  GppGood as GppGoodIcon,
  QrCode as QrCodeIcon,
} from "@mui/icons-material";
import {
  Box,
  Button,
  Card,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
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

export default function TOTPSection({ user }: { user: UserData }) {
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
        setCode("");
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
    <Stack spacing={3}>
      {/* ヘッダー部分 */}
      <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
        <Typography variant="h6" sx={{ component: "h4" }}>
          二段階認証（TOTP）
        </Typography>
        <Divider sx={{ flexGrow: 1, ml: 2 }} />
      </Stack>

      {/* ステータスとアクションボタン */}
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={3}
        sx={{ alignItems: { xs: "flex-start", sm: "center" } }}
      >
        <Stack direction="row" spacing={2} sx={{ alignItems: "center" }}>
          <Typography variant="body1" sx={{ fontWeight: "medium" }}>
            現在の状態
          </Typography>
          {localEnabled ? (
            <Chip
              icon={<GppGoodIcon />}
              label="有効"
              color="success"
              variant="filled"
              size="small"
              sx={{ fontWeight: "bold" }}
            />
          ) : (
            <Chip
              icon={<GppBadIcon />}
              label="無効"
              color="default"
              variant="filled"
              size="small"
            />
          )}
        </Stack>

        {!localEnabled && !genResult?.secret && (
          <Button
            variant="contained"
            color="primary"
            onClick={() => setOpenDialog(true)}
            disableElevation
          >
            セットアップを開始
          </Button>
        )}

        {localEnabled && (
          <Button
            variant="outlined"
            color="error"
            onClick={() => setOpenDisableDialog(true)}
          >
            無効化する
          </Button>
        )}
      </Stack>

      {/* セットアップ領域（QRコード表示部） */}
      {genResult?.secret && isFinished && !localEnabled && (
        <Card
          variant="outlined"
          sx={{ p: 3, bgcolor: "info.50", borderColor: "info.200" }}
        >
          <Stack
            spacing={3}
            direction={{ xs: "column", md: "row" }}
            sx={{ alignItems: "center" }}
          >
            {uri && (
              <Box
                sx={{
                  p: 2,
                  bgcolor: "white",
                  borderRadius: 2,
                  boxShadow: 1,
                  display: "flex",
                }}
              >
                <QRCodeCanvas value={uri} size={160} />
              </Box>
            )}

            <Stack spacing={2} sx={{ flexGrow: 1, width: "100%" }}>
              <Box>
                <Stack
                  direction="row"
                  spacing={1}
                  sx={{ alignItems: "center", mb: 1 }}
                >
                  <QrCodeIcon color="info" fontSize="small" />
                  <Typography
                    variant="subtitle2"
                    sx={{ fontWeight: "bold", color: "info.dark" }}
                  >
                    認証アプリでQRコードをスキャン
                  </Typography>
                </Stack>
                <Typography variant="body2" color="text.secondary" gutterBottom>
                  Google Authenticator
                  などの認証アプリを使用してQRコードを読み取ってください。カメラが使えない場合は、以下のシークレットを手動で入力してください。
                </Typography>
                <Typography
                  variant="body2"
                  sx={{
                    mt: 1,
                    p: 1,
                    bgcolor: "white",
                    borderRadius: 1,
                    fontFamily: "monospace",
                    border: "1px solid",
                    borderColor: "grey.300",
                    wordBreak: "break-all",
                  }}
                >
                  シークレット: <strong>{genResult.secret}</strong>
                </Typography>
              </Box>

              <Box component="form" action={verifyAction} sx={{ mt: 1 }}>
                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  spacing={2}
                  sx={{ alignItems: "flex-start" }}
                >
                  <TextField
                    label="6桁の認証コード"
                    name="code"
                    value={code}
                    onChange={(e) => setCode(e.target.value)}
                    size="small"
                    fullWidth
                    sx={{ bgcolor: "white", borderRadius: 1 }}
                  />
                  <Button
                    type="submit"
                    variant="contained"
                    color="primary"
                    disabled={!code}
                    sx={{ whiteSpace: "nowrap", py: 1 }}
                    disableElevation
                  >
                    有効化
                  </Button>
                </Stack>
              </Box>
            </Stack>
          </Stack>
        </Card>
      )}

      {/* セットアップ用パスワードダイアログ */}
      <Dialog
        open={openDialog}
        onClose={() => setOpenDialog(false)}
        fullWidth
        maxWidth="xs"
      >
        <DialogTitle>二段階認証のセットアップ</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }} variant="body2">
            セキュリティ上の理由から、続行するには現在のパスワードを入力してください。
          </DialogContentText>
          <Box component="form" action={genAction} id="totp-gen-form">
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
          <DialogActions
            sx={{
              flexDirection: { xs: "column", sm: "row" },
            }}
          >
            <Button
              onClick={() => setOpenDialog(false)}
              color="inherit"
              sx={{ width: { xs: "100%", sm: "auto" } }}
            >
              キャンセル
            </Button>
            <Button
              type="submit"
              form="totp-gen-form"
              variant="contained"
              disableElevation
              sx={{ width: { xs: "100%", sm: "auto" } }}
            >
              確認して生成
            </Button>
          </DialogActions>
        </DialogContent>
      </Dialog>

      {/* 無効化用パスワードダイアログ */}
      <Dialog
        open={openDisableDialog}
        onClose={() => setOpenDisableDialog(false)}
        fullWidth
        maxWidth="xs"
      >
        <DialogTitle sx={{ fontWeight: "bold", color: "error.main" }}>
          二段階認証の無効化
        </DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ mb: 2 }} variant="body2">
            アカウントのセキュリティレベルが低下します。続行するには現在のパスワードを入力してください。
          </DialogContentText>
          <Box component="form" action={disableAction} id="totp-disable-form">
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
          <DialogActions
            sx={{
              flexDirection: { xs: "column", sm: "row" },
            }}
          >
            <Button
              onClick={() => setOpenDisableDialog(false)}
              color="inherit"
              sx={{ width: { xs: "100%", sm: "auto" } }}
            >
              キャンセル
            </Button>
            <Button
              type="submit"
              form="totp-disable-form"
              variant="contained"
              color="error"
              disableElevation
              sx={{ width: { xs: "100%", sm: "auto" } }}
            >
              無効化する
            </Button>
          </DialogActions>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
