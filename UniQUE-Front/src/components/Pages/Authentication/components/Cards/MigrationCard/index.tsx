import { FormHelperText } from "@mui/material";
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
import { submitMigration } from "../../actions/migration";
import { SitemarkIcon } from "../../CustomIcons";
import { Card } from "../Base";

export default function MigrationCard() {
  const initialState = useInitialFormState();

  const [emailError, setEmailError] = React.useState(false);
  const [emailErrorMessage, setEmailErrorMessage] = React.useState(
    "@uniproject.jpで終わるメールアドレスを入力してください。",
  );
  const [externalEmailError, setExternalEmailError] = React.useState(false);
  const [externalEmailErrorMessage, setExternalEmailErrorMessage] =
    React.useState(
      "UniProjectに登録している@uniproject.jp以外のメールアドレスを入力してください。不明な場合はお問合せください。",
    );
  const [passwordError, setPasswordError] = React.useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = React.useState(
    "お好きなパスワードを8文字以上で設定してください。",
  );
  const [confirmPasswordError, setConfirmPasswordError] = React.useState(false);
  const [confirmPasswordErrorMessage, setConfirmPasswordErrorMessage] =
    React.useState("");
  const [agreeTosError, setAgreeTosError] = React.useState(false);
  const [agreeTosErrorMessage, setAgreeTosErrorMessage] = React.useState("");

  const validateInputs = () => {
    const email = document.getElementById("email") as HTMLInputElement;
    const password = document.getElementById("password") as HTMLInputElement;
    const confirmPassword = document.getElementById(
      "confirm_password",
    ) as HTMLInputElement;

    let isValid = true;

    if (!email.value || !/\S+@uniproject.jp/.test(email.value)) {
      setEmailError(true);
      setEmailErrorMessage("有効なメールアドレスを入力してください。");
      isValid = false;
    } else {
      setEmailError(false);
      setEmailErrorMessage("");
    }

    if (!password.value || password.value.length < 8) {
      setPasswordError(true);
      setPasswordErrorMessage("パスワードは8文字以上で設定してください。");
      isValid = false;
    } else {
      setPasswordError(false);
      setPasswordErrorMessage("");
    }

    if (confirmPassword.value !== password.value) {
      setConfirmPasswordError(true);
      setConfirmPasswordErrorMessage("Passwords do not match.");
      isValid = false;
    } else {
      setConfirmPasswordError(false);
      setConfirmPasswordErrorMessage("");
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

    const externalEmail = document.getElementById(
      "external_email",
    ) as HTMLInputElement;
    if (
      !externalEmail.value ||
      !/\S+@\S+\.\S+/.test(externalEmail.value) ||
      /@uniproject\.jp\s*$/.test(externalEmail.value)
    ) {
      setExternalEmailError(true);
      setExternalEmailErrorMessage(
        "有効なメールアドレスを入力してください。UniProjectに登録しているメールアドレスは使用できません。",
      );
      isValid = false;
    } else {
      setExternalEmailError(false);
      setExternalEmailErrorMessage(
        "UniProjectに登録しているGmailやOutlookなどのメールアドレスを入力してください。",
      );
    }

    return isValid;
  };

  return (
    <Card variant="outlined">
      <Box sx={{ display: { xs: "flex", md: "none" } }}>
        <SitemarkIcon />
      </Box>
      <Typography
        component="h1"
        variant="h4"
        sx={{ width: "100%", fontSize: "clamp(2rem, 10vw, 2.15rem)" }}
      >
        アカウント移行
      </Typography>
      <Typography sx={{ textAlign: "left" }}>
        アカウントをお持ちですか？{" "}
        <span>
          <Link href="/signin" variant="body2" sx={{ alignSelf: "center" }}>
            サインイン
          </Link>
        </span>
      </Typography>
      <Typography sx={{ textAlign: "left" }}>
        既存メンバーではありませんか？{" "}
        <span>
          <Link href="/signup" variant="body2" sx={{ alignSelf: "center" }}>
            登録申請
          </Link>
        </span>
      </Typography>
      <Box
        component="form"
        noValidate
        sx={{ display: "flex", flexDirection: "column", width: "100%", gap: 2 }}
        action={submitMigration}
      >
        <FormControl>
          <FormLabel htmlFor="name">お名前</FormLabel>
          <TextField
            helperText={
              "ニックネームでも構いません。お好きな表示名を設定してください。"
            }
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
          <FormLabel htmlFor="email">内部メールアドレス</FormLabel>
          <TextField
            error={emailError}
            helperText={emailErrorMessage}
            id="email"
            type="email"
            name="email"
            placeholder="your@uniproject.jp"
            defaultValue={initialState?.email}
            autoComplete="email"
            autoFocus
            required
            fullWidth
            variant="outlined"
            color={emailError ? "error" : "primary"}
          />
        </FormControl>
        <FormControl>
          <FormLabel htmlFor="external_email">外部メールアドレス</FormLabel>
          <TextField
            error={externalEmailError}
            helperText={externalEmailErrorMessage}
            id="external_email"
            type="email"
            name="external_email"
            placeholder="your@example.com"
            defaultValue={initialState?.externalEmail}
            autoComplete="email"
            autoFocus
            required
            fullWidth
            variant="outlined"
            color={externalEmailError ? "error" : "primary"}
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
            <FormLabel htmlFor="confirm_password">パスワード（確認）</FormLabel>
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
    </Card>
  );
}
