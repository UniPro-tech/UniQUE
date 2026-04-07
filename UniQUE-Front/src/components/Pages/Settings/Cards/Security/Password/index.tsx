"use client";
import {
  Button,
  Divider,
  FormHelperText,
  Link,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { enqueueSnackbar } from "notistack";
import { useActionState, useCallback, useEffect, useState } from "react";
import type { UserData } from "@/classes/types/User";
import type { FormStatus } from "../../Base";
import { updateSettings } from "./action";

export default function PasswordSection({
  user,
  csrfToken,
  handleClickOpen,
}: {
  user: UserData;
  csrfToken: string;
  handleClickOpen: () => void;
}) {
  const [lastResult, action] = useActionState(
    updateSettings,
    null as null | FormStatus,
  );
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmNewPassword, setConfirmNewPassword] = useState("");
  const [newPasswordError, setNewPasswordError] = useState("");
  const [confirmPasswordError, setConfirmPasswordError] = useState("");

  const validateNewPassword = useCallback((value: string) => {
    if (value.length === 0) {
      setNewPasswordError("");
      return true;
    }
    if (value.length < 8) {
      setNewPasswordError("パスワードは8文字以上で入力してください。");
      return false;
    }
    setNewPasswordError("");
    return true;
  }, []);

  const validateConfirmPassword = useCallback((value: string, pw: string) => {
    if (value.length === 0) {
      setConfirmPasswordError("");
      return true;
    }
    if (value !== pw) {
      setConfirmPasswordError("パスワードが一致しません。");
      return false;
    }
    setConfirmPasswordError("");
    return true;
  }, []);
  useEffect(() => {
    if (lastResult) {
      enqueueSnackbar(lastResult.message, { variant: lastResult.status });
    }
  }, [lastResult]);
  return (
    <Stack
      component="form"
      action={action}
      spacing={2}
      onSubmit={(e) => {
        // クライアント側検証。不正なら送信を止める。
        const ok1 = validateNewPassword(newPassword);
        const ok2 = validateConfirmPassword(confirmNewPassword, newPassword);
        if (!ok1 || !ok2) {
          e.preventDefault();
          enqueueSnackbar("入力内容を確認してください。", { variant: "error" });
        }
      }}
    >
      <Stack direction="row" alignItems="center" spacing={2}>
        <Typography variant="h6" component={"h4"}>
          パスワードの変更
        </Typography>
        <Divider sx={{ flexGrow: 1 }} />
      </Stack>
      <input type="hidden" name="csrfToken" value={csrfToken} />
      <input type="hidden" name="id" value={user.id!} />
      <TextField
        label="現在のパスワード"
        name="current_password"
        type="password"
        fullWidth
        value={currentPassword}
        onChange={(e) => setCurrentPassword(e.target.value)}
      />
      <TextField
        label="新しいパスワード"
        name="new_password"
        type="password"
        fullWidth
        value={newPassword}
        onChange={(e) => {
          setNewPassword(e.target.value);
          validateNewPassword(e.target.value);
          // 再チェック（確認欄）
          validateConfirmPassword(confirmNewPassword, e.target.value);
        }}
        error={newPasswordError.length > 0}
        helperText={
          newPasswordError || "8文字以上で、安全なパスワードを決めてください"
        }
      />
      <Stack>
        <TextField
          label="新しいパスワード（確認）"
          name="confirm_new_password"
          type="password"
          fullWidth
          value={confirmNewPassword}
          onChange={(e) => {
            setConfirmNewPassword(e.target.value);
            validateConfirmPassword(e.target.value, newPassword);
          }}
          error={confirmPasswordError.length > 0}
          helperText={
            confirmPasswordError ||
            "上記で入力したパスワードをもう一度ご入力ください"
          }
        />
        <FormHelperText>
          パスワードを忘れた場合は、{" "}
          <Link
            onClick={handleClickOpen}
            underline="hover"
            style={{ cursor: "pointer" }}
          >
            こちら
          </Link>
        </FormHelperText>
      </Stack>
      <Button variant="contained" fullWidth type="submit">
        保存
      </Button>
    </Stack>
  );
}
