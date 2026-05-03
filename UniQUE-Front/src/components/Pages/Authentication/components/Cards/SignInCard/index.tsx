"use client";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import FormControl from "@mui/material/FormControl";
import FormControlLabel from "@mui/material/FormControlLabel";
import FormLabel from "@mui/material/FormLabel";
import Link from "@mui/material/Link";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { useRouter } from "next/navigation";
import { SnackbarProvider } from "notistack";
import * as React from "react";
import { useInitialFormState, useRedirectTo } from "../../../Client";
import { submitSignIn } from "../../actions/signIn";
import { SitemarkIcon } from "../../CustomIcons";
import ForgotPassword from "../../ForgotPassword";
import { Card } from "../Base";

export default function SignUpCard() {
  const initialState = useInitialFormState();
  const redirectTo = useRedirectTo();

  const [passwordError, setPasswordError] = React.useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = React.useState("");

  const [forgotPasswordOpen, setForgotPasswordOpen] = React.useState(false);

  const [inProgress, setInProgress] = React.useState(false);

  const handleClickOpen = () => {
    setForgotPasswordOpen(true);
  };

  const handleClose = () => {
    setForgotPasswordOpen(false);
  };

  const validateInputs = () => {
    setInProgress(true);
    const password = document.getElementById("password") as HTMLInputElement;

    let isValid = true;

    if (!password.value) {
      setPasswordError(true);
      setPasswordErrorMessage("パスワードを入力してください。");
      isValid = false;
    } else {
      setPasswordError(false);
      setPasswordErrorMessage("");
    }

    if (!isValid) setInProgress(false);

    return isValid;
  };

  const router = useRouter();

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
        サインイン
      </Typography>
      <Typography sx={{ textAlign: "left" }}>
        メンバーではありませんか？{" "}
        <span>
          <Link href="/signup" variant="body2" sx={{ alignSelf: "center" }}>
            サインアップ
          </Link>
        </span>
      </Typography>
      <Typography sx={{ textAlign: "left" }}>
        既存メンバーですか？{" "}
        <span>
          <Link href="/migrate" variant="body2" sx={{ alignSelf: "center" }}>
            アカウント移行
          </Link>
        </span>
      </Typography>
      <Box
        component="form"
        noValidate
        sx={{ display: "flex", flexDirection: "column", width: "100%", gap: 2 }}
        action={async (data: FormData) => {
          setInProgress(true);
          const result = await submitSignIn(data);
          if (result.redirectTo) {
            router.push(result.redirectTo);
          }
          setInProgress(false);
        }}
      >
        <input
          type="hidden"
          name="redirectTo"
          value={redirectTo ? redirectTo : "/dashboard"}
        />
        <FormControl>
          <FormLabel htmlFor="username">カスタムID</FormLabel>
          <TextField
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
            color={"primary"}
          />
        </FormControl>
        <FormControl>
          <Box sx={{ display: "flex", justifyContent: "space-between" }}>
            <FormLabel htmlFor="password">パスワード</FormLabel>
            <Link
              component="button"
              type="button"
              onClick={handleClickOpen}
              variant="body2"
              disabled={inProgress}
              sx={{ alignSelf: "baseline" }}
            >
              パスワードをお忘れですか？
            </Link>
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
        <FormControlLabel
          control={<Checkbox value="remember" color="primary" />}
          label={"ログイン状態を保持する"}
        />
        <Button
          type="submit"
          fullWidth
          variant="contained"
          onSubmit={validateInputs}
          color={"primary"}
          disabled={inProgress}
        >
          {!inProgress ? "サインイン" : "サインイン中..."}
        </Button>
      </Box>
      <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
        <ForgotPassword open={forgotPasswordOpen} handleClose={handleClose} />
      </SnackbarProvider>
    </Card>
  );
}
