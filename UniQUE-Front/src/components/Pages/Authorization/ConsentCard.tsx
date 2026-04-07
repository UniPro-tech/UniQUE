"use client";

import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import MuiCard from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Divider from "@mui/material/Divider";
import Link from "@mui/material/Link";
import Stack from "@mui/material/Stack";
import type { Theme } from "@mui/material/styles";
import Typography from "@mui/material/Typography";
import type { ApplicationData } from "@/classes/Application";
import type { UserData } from "@/classes/types/User";

export default function ConsentCard(props: {
  app: ApplicationData;
  user?: UserData;
  scope: string;
  jwt: string;
  redirect_uri: string;
  state?: string;
  action?: string;
  auth_request_id?: string;
}) {
  const {
    app,
    user,
    scope,
    jwt,
    redirect_uri,
    state,
    action,
    auth_request_id,
  } = props;
  const scopes = scope ? scope.split(" ") : [];
  // `action` should be provided by the server page (uses AUTH_API_URL). Fallback to relative endpoint.
  const formAction = action ?? "/auth";

  const denyHref = `${redirect_uri}?error=access_denied${
    state ? `&state=${encodeURIComponent(state)}` : ""
  }`;

  return (
    <MuiCard
      variant="outlined"
      sx={(theme: Theme) => ({
        display: "flex",
        flexDirection: "column",
        alignSelf: "center",
        width: "100%",
        padding: theme.spacing(4),
        gap: theme.spacing(2),
        boxShadow:
          "hsla(220, 30%, 5%, 0.05) 0px 5px 15px 0px, hsla(220, 25%, 10%, 0.05) 0px 15px 35px -5px",
        [theme.breakpoints.up("sm")]: {
          width: "600px",
        },
        [theme.breakpoints.down("sm")]: {
          padding: theme.spacing(2),
        },
      })}
    >
      <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
        <Avatar sx={{ width: 56, height: 56, bgcolor: "primary.main" }}>
          {app.name ? app.name.charAt(0).toUpperCase() : "A"}
        </Avatar>
        <Box>
          <Typography component="h1" variant="h6">
            {app.name} にアクセスを許可しますか？
          </Typography>
          <Typography variant="caption" color="text.secondary">
            クライアントID: {app.id || "-"}
          </Typography>
        </Box>
      </Box>

      <Box sx={{ mt: 1 }}>
        {app.websiteUrl && (
          <Typography variant="body2" color="text.secondary">
            ウェブサイト:
            <Link
              href={app.websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
              underline="hover"
              sx={{ ml: 0.5 }}
            >
              {app.websiteUrl}
            </Link>
          </Typography>
        )}

        {app.privacyPolicyUrl && (
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            プライバシーポリシー:
            <Link
              href={app.privacyPolicyUrl}
              target="_blank"
              rel="noopener noreferrer"
              underline="hover"
              sx={{ ml: 0.5 }}
            >
              表示する
            </Link>
          </Typography>
        )}
      </Box>

      <Divider sx={{ my: 1.5 }} />

      <CardContent sx={{ px: 0 }}>
        <Box sx={{ mb: 1 }}>
          <Typography variant="body2">要求されている権限</Typography>
          <Stack spacing={1} sx={{ mt: 1 }}>
            {scopes.length === 0 && (
              <Typography variant="body2" color="text.secondary">
                （なし）
              </Typography>
            )}
            {scopes.map((s) => (
              <Box key={s} sx={{ display: "flex", gap: 1 }}>
                <Typography variant="body2">•</Typography>
                <Typography variant="body2">{s}</Typography>
              </Box>
            ))}
          </Stack>
        </Box>

        <Box>
          <Typography variant="body2">ログイン中のユーザー</Typography>
          <Typography variant="subtitle2">
            {user?.customId ?? user?.email ?? "不明なユーザー"}
          </Typography>
        </Box>
      </CardContent>

      <CardActions sx={{ px: 0, mt: 1 }}>
        <Box sx={{ display: "flex", gap: 2, width: "100%", flexWrap: "wrap" }}>
          <Box
            component="form"
            action={formAction}
            method="post"
            sx={{ flex: 1 }}
          >
            <input type="hidden" name="session_jwt" value={jwt} />
            <input
              type="hidden"
              name="auth_request_id"
              value={auth_request_id}
            />
            {state && <input type="hidden" name="state" value={state} />}
            <Button type="submit" variant="contained" color="primary" fullWidth>
              許可する
            </Button>
          </Box>

          <Box sx={{ flexBasis: 200 }}>
            <Link href={denyHref} underline="none">
              <Button variant="outlined" color="inherit" fullWidth>
                拒否する
              </Button>
            </Link>
          </Box>
        </Box>
      </CardActions>
    </MuiCard>
  );
}
