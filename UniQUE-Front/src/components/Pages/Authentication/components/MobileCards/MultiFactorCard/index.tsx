"use client";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import FormLabel from "@mui/material/FormLabel";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useInitialFormState, useRedirectTo } from "../../../Client";
import { submitSignIn } from "../../actions/signIn";
import { SitemarkIcon } from "../../CustomIcons";
import { Card } from "../Base";

export default function MultiFactorCard() {
  const initialState = useInitialFormState();
  const redirectTo = useRedirectTo();

  const [inProgress, setInProgress] = useState(false);

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
        2段階認証
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
        <input type="hidden" name="redirectTo" value={redirectTo} />
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
            disabled
            fullWidth
            variant="outlined"
            color={"primary"}
          />
        </FormControl>
        <input type="hidden" name="username" value={initialState?.username} />
        <FormControl>
          <Box sx={{ display: "flex", justifyContent: "space-between" }}>
            <FormLabel htmlFor="password">2段階認証コード</FormLabel>
          </Box>
          <TextField
            name="code"
            placeholder="••••••"
            type="text"
            id="code"
            autoComplete="one-time-code"
            autoFocus
            required
            fullWidth
            variant="outlined"
            color={"primary"}
            helperText="認証アプリで生成されたコードを入力してください。"
          />
        </FormControl>
        <input
          type="hidden"
          name="remember"
          value={initialState?.rememberMe ? "on" : "off"}
        />
        <Button
          type="submit"
          variant="contained"
          fullWidth
          disabled={inProgress}
        >
          {!inProgress ? "サインイン" : "サインイン中..."}
        </Button>
      </Box>
    </Card>
  );
}
