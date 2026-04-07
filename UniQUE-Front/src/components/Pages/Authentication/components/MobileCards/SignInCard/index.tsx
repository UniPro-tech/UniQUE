import { Divider } from "@mui/material";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import FormControl from "@mui/material/FormControl";
import FormControlLabel from "@mui/material/FormControlLabel";
import FormLabel from "@mui/material/FormLabel";
import Link from "@mui/material/Link";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { RedirectType, redirect } from "next/navigation";
import * as React from "react";
import { useInitialFormState, useRedirectTo } from "../../../Client";
import { submitSignIn } from "../../actions/signIn";
import { SitemarkIcon } from "../../CustomIcons";
import { Card } from "../Base";

export default function SignInCard() {
  const initialState = useInitialFormState();
  const redirectTo = useRedirectTo();

  const [passwordError, setPasswordError] = React.useState(false);
  const [passwordErrorMessage, setPasswordErrorMessage] = React.useState("");

  const validateInputs = () => {
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
        サインイン
      </Typography>
      <Box
        component="form"
        noValidate
        sx={{ display: "flex", flexDirection: "column", width: "100%", gap: 2 }}
        action={async (data: FormData) => {
          const result = await submitSignIn(data);
          if (result.redirectTo) {
            redirect(result.redirectTo, RedirectType.push);
          }
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
          onClick={validateInputs}
        >
          サインイン
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
          既存メンバーですか？{" "}
          <Link href="/migrate" variant="body2" sx={{ alignSelf: "center" }}>
            アカウント移行
          </Link>
        </Typography>
      </Box>
    </Card>
  );
}
