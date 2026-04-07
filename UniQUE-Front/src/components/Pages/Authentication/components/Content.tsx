"use client";
import BadgeIcon from "@mui/icons-material/Badge";
import CachedIcon from "@mui/icons-material/Cached";
import EmojiPeopleIcon from "@mui/icons-material/EmojiPeople";
import HubIcon from "@mui/icons-material/Hub";
import KeyIcon from "@mui/icons-material/Key";
import PublishedWithChangesIcon from "@mui/icons-material/PublishedWithChanges";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { AuthorizationPageMode, useAuthPageMode } from "../Client";
import { SitemarkIcon } from "./CustomIcons";

const SignInItems = [
  {
    icon: <HubIcon sx={{ color: "text.secondary" }} />,
    title: "ひとつのIDで、すべてをつなぐ",
    description: (
      <>
        UniQUEひとつで、UniProjectの
        <wbr />
        すべての
        <wbr />
        サービスを
        <wbr />
        利用できます。
      </>
    ),
  },
  {
    icon: <KeyIcon sx={{ color: "text.secondary" }} />,
    title: "シンプルでセキュア",
    description: (
      <>
        OIDC/OAuth2.0対応。
        <wbr />
        安全でスマートなログインを。
      </>
    ),
  },
  {
    icon: <BadgeIcon sx={{ color: "text.secondary" }} />,
    title: "メンバーであることを証明",
    description: (
      <>
        本当に必要な人だけが
        <wbr />
        アクセスできるように、
        <wbr />
        メンバー管理を簡単に。
      </>
    ),
  },
  {
    icon: <PublishedWithChangesIcon sx={{ color: "text.secondary" }} />,
    title: "リアルタイムで正確なデータを",
    description: (
      <>
        簡単に素早く情報を更新でき、
        <wbr />
        常に
        <wbr />
        最新の
        <wbr />
        状態を
        <wbr />
        保ちます。
      </>
    ),
  },
];

export default function Content() {
  const mode = useAuthPageMode();
  return (
    <Stack
      sx={{
        flexDirection: "column",
        alignSelf: "center",
        gap: 4,
        maxWidth: 450,
      }}
    >
      <Box sx={{ display: { xs: "none", md: "flex" } }}>
        <SitemarkIcon />
      </Box>
      {(() => {
        switch (mode) {
          case AuthorizationPageMode.SignUp:
            return (
              <Stack direction="row" gap={2}>
                <EmojiPeopleIcon sx={{ color: "text.secondary" }} />
                <Box display={"block"} width={"100%"}>
                  <Typography gutterBottom sx={{ fontWeight: "medium" }}>
                    UniProjectへようこそ！
                  </Typography>
                  <Typography
                    variant="body2"
                    sx={{
                      color: "text.secondary",
                      wordBreak: "keep-all",
                      width: "100%",
                    }}
                  >
                    UniProjectに興味をお持ちいただき、ありがとうございます。
                    <wbr />
                    参加するには、当サービス「UniQUE」への登録が必要です。
                    <wbr />
                    右側のフォームに必要事項を入力して、登録申請を送信してください。
                    <wbr />
                    たくさんのご参加をお待ちしています。
                  </Typography>
                </Box>
              </Stack>
            );
          case AuthorizationPageMode.SignIn:
          case AuthorizationPageMode.MFA:
            return SignInItems.map((item) => (
              <Stack key={item.title} direction="row" sx={{ gap: 2 }}>
                {item.icon}
                <Box sx={{ display: "block", width: "100%" }}>
                  <Typography gutterBottom sx={{ fontWeight: "medium" }}>
                    {item.title}
                  </Typography>
                  <Typography
                    variant="body2"
                    sx={{
                      color: "text.secondary",
                      wordBreak: "keep-all",
                      width: "100%",
                    }}
                  >
                    {item.description}
                  </Typography>
                </Box>
              </Stack>
            ));
          case AuthorizationPageMode.Migration:
            return (
              <Stack direction="row" gap={2}>
                <CachedIcon sx={{ color: "text.secondary" }} />
                <Box display={"block"} width={"100%"}>
                  <Typography gutterBottom sx={{ fontWeight: "medium" }}>
                    UniProjectアカウント移行手続き
                  </Typography>
                  <Typography
                    variant="body2"
                    sx={{
                      color: "text.secondary",
                      wordBreak: "keep-all",
                      width: "100%",
                    }}
                  >
                    既存のUniProjectアカウントを新しい認証システムに移行します。
                    <wbr />
                    右側のフォームに必要事項を入力して、移行手続きを完了させてください。
                  </Typography>
                </Box>
              </Stack>
            );
        }
      })()}
    </Stack>
  );
}
