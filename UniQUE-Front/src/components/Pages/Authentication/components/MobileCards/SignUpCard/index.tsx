import { Divider, FormHelperText } from "@mui/material";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import FormControl from "@mui/material/FormControl";
import FormControlLabel from "@mui/material/FormControlLabel";
import FormLabel from "@mui/material/FormLabel";
import Link from "@mui/material/Link";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import * as React from "react";
import { useInitialFormState } from "../../../Client";
import { submitSignUp } from "../../actions/signUp";
import { SitemarkIcon } from "../../CustomIcons";
import { Card, SignInContainer } from "../Base";

export default function SignUpCard() {
  const initialState = useInitialFormState();

  const [emailError, setEmailError] = React.useState(false);
  const [emailErrorMessage, setEmailErrorMessage] = React.useState("");
  const [passwordError, setPasswordError] = React.useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = React.useState(
    "8文字以上のお好きなパスワードを設定してください。",
  );
  const [confirmPasswordError, setConfirmPasswordError] = React.useState(false);
  const [confirmPasswordErrorMessage, setConfirmPasswordErrorMessage] =
    React.useState("");
  const [usernameError, setUsernameError] = React.useState(false);
  const [usernameErrorMessage, setUsernameErrorMessage] = React.useState(
    "カスタムIDは半角英数字とアンダースコアのみで設定してください。",
  );
  const [agreeTosError, setAgreeTosError] = React.useState(false);
  const [agreeTosErrorMessage, setAgreeTosErrorMessage] = React.useState("");

  const validateInputs = () => {
    const email = document.getElementById("email") as HTMLInputElement;
    const password = document.getElementById("password") as HTMLInputElement;
    const confirmPassword = document.getElementById(
      "confirm_password",
    ) as HTMLInputElement;
    const username = document.getElementById("username") as HTMLInputElement;

    let isValid = true;

    if (!email.value || !/\S+@\S+\.\S+/.test(email.value)) {
      setEmailError(true);
      setEmailErrorMessage("有効なメールアドレスを入力してください。");
      isValid = false;
    } else {
      setEmailError(false);
      setEmailErrorMessage("");
    }

    if (!password.value || password.value.length < 6) {
      setPasswordError(true);
      setPasswordErrorMessage("パスワードは8文字以上で設定してください。");
      isValid = false;
    } else {
      setPasswordError(false);
      setPasswordErrorMessage("");
    }

    if (confirmPassword.value !== password.value) {
      setConfirmPasswordError(true);
      setConfirmPasswordErrorMessage("パスワードが一致しません。");
      isValid = false;
    } else {
      setConfirmPasswordError(false);
      setConfirmPasswordErrorMessage("");
    }

    if (!username.value || !/^[a-zA-Z0-9_]+$/.test(username.value)) {
      // カスタムIDが空であるか、半角英数字とアンダースコア以外の文字が含まれている場合はエラー
      setUsernameError(true);
      setUsernameErrorMessage(
        "カスタムIDは半角英数字とアンダースコアのみで設定してください。",
      );
      isValid = false;
    } else {
      setUsernameError(false);
      setUsernameErrorMessage(
        "カスタムIDは半角英数字とアンダースコアのみで設定してください。",
      );
    }

    // もし数字や_のみであればエラー
    if (/^[0-9_]+$/.test(username.value)) {
      setUsernameError(true);
      setUsernameErrorMessage("カスタムIDは半角英字も含めて設定してください。");
      isValid = false;
    }

    // 先頭の文字が数字や_であればエラー
    if (/^[0-9_]/.test(username.value)) {
      setUsernameError(true);
      setUsernameErrorMessage("カスタムIDの先頭は半角英字で設定してください。");
      isValid = false;
    }

    // 最後の文字が_であればエラー
    if (/_$/.test(username.value)) {
      setUsernameError(true);
      setUsernameErrorMessage(
        "カスタムIDの最後は半角英字または数字で設定してください。",
      );
      isValid = false;
    }

    const agreeTos = document.querySelector(
      'input[name="agreeTos"]',
    ) as HTMLInputElement;
    if (!agreeTos.checked) {
      setAgreeTosError(true);
      setAgreeTosErrorMessage(
        "利用規約、プライバシーポリシー、サークル規約に同意してください。",
      );
      isValid = false;
    } else {
      setAgreeTosError(false);
      setAgreeTosErrorMessage(
        "利用規約、プライバシーポリシー、サークル規約に同意してください。",
      );
    }

    return isValid;
  };

  return (
    <SignInContainer direction="column" justifyContent="space-between">
      <Card variant="outlined">
        <Box sx={{ display: { xs: "flex", md: "none" } }}>
          <SitemarkIcon />
        </Box>
        <Typography
          component="h1"
          variant="h4"
          sx={{ width: "100%", fontSize: "clamp(2rem, 10vw, 2.15rem)" }}
        >
          サインアップ
        </Typography>
        <Box
          component="form"
          noValidate
          sx={{
            display: "flex",
            flexDirection: "column",
            width: "100%",
            gap: 2,
          }}
          action={submitSignUp}
        >
          <FormControl>
            <FormLabel htmlFor="name">お名前</FormLabel>
            <TextField
              helperText={"ニックネームでも構いません"}
              id="name"
              type="text"
              name="name"
              placeholder="your name"
              defaultValue={initialState?.name}
              autoComplete="name"
              autoFocus
              required
              fullWidth
              variant="outlined"
              color={emailError ? "error" : "primary"}
            />
          </FormControl>
          <FormControl>
            <FormLabel htmlFor="username">カスタムID</FormLabel>
            <TextField
              error={usernameError}
              helperText={usernameErrorMessage}
              id="username"
              type="text"
              name="username"
              placeholder="your_username"
              defaultValue={initialState?.username}
              autoComplete="username"
              autoFocus
              required
              fullWidth
              variant="outlined"
              color={usernameError ? "error" : "primary"}
            />
          </FormControl>
          <FormControl>
            <FormLabel htmlFor="external_email">メールアドレス</FormLabel>
            <TextField
              error={emailError}
              helperText={emailErrorMessage}
              id="external_email"
              type="email"
              name="external_email"
              placeholder="your@email.com"
              defaultValue={initialState?.externalEmail}
              autoComplete="email"
              autoFocus
              required
              fullWidth
              variant="outlined"
              color={emailError ? "error" : "primary"}
            />
          </FormControl>
          <FormControl>
            <Box sx={{ display: "flex", justifyContent: "space-between" }}>
              <FormLabel htmlFor="password">パスワード</FormLabel>
            </Box>
            <TextField
              error={passwordError}
              helperText={passwordErrorMessage}
              name="password"
              placeholder="••••••"
              type="password"
              id="password"
              autoComplete="current-password"
              autoFocus
              required
              fullWidth
              variant="outlined"
              color={passwordError ? "error" : "primary"}
            />
          </FormControl>
          <FormControl>
            <Box sx={{ display: "flex", justifyContent: "space-between" }}>
              <FormLabel htmlFor="confirm_password">
                パスワード（確認）
              </FormLabel>
            </Box>
            <TextField
              error={confirmPasswordError}
              helperText={confirmPasswordErrorMessage}
              name="confirm_password"
              placeholder="••••••"
              type="password"
              id="confirm_password"
              autoFocus
              required
              fullWidth
              variant="outlined"
              color={confirmPasswordError ? "error" : "primary"}
            />
          </FormControl>
          <FormControlLabel
            control={<Checkbox value="agreeTos" color="primary" />}
            label={
              <>
                <Link href="/terms" target="_blank">
                  利用規約
                </Link>
                と
                <Link href="/privacy" target="_blank">
                  プライバシーポリシー
                </Link>
                、
                <Link href="/club_statute" target="_blank">
                  サークル規約
                </Link>
                に同意します
              </>
            }
          />
          <FormHelperText error={agreeTosError}>
            {agreeTosErrorMessage}
          </FormHelperText>
          <Button
            type="submit"
            fullWidth
            variant="contained"
            onClick={validateInputs}
          >
            サインアップ
          </Button>
        </Box>
        <Divider>or</Divider>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Typography sx={{ textAlign: "center" }}>
            メンバーではありませんか？{" "}
            <Link href="/signup" variant="body2" sx={{ alignSelf: "center" }}>
              登録申請
            </Link>
          </Typography>
        </Box>
        <Divider>or</Divider>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Typography sx={{ textAlign: "center" }}>
            メンバーではありませんか？{" "}
            <Link href="/signup" variant="body2" sx={{ alignSelf: "center" }}>
              登録申請
            </Link>
          </Typography>
        </Box>
        <Divider>or</Divider>
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Typography sx={{ textAlign: "center" }}>
            既存メンバーですか？{" "}
            <Link href="/migrate" variant="body2" sx={{ alignSelf: "center" }}>
              アカウント移行
            </Link>
          </Typography>
        </Box>
      </Card>
    </SignInContainer>
  );
}
