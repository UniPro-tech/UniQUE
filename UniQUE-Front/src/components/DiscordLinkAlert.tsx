"use client";
import CloseIcon from "@mui/icons-material/Close";
import {
  Alert,
  AlertTitle,
  Button,
  Collapse,
  IconButton,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

export default function DiscordLinkAlert() {
  const [isDiscordLinked, setIsDiscordLinked] = useState<boolean | null>(null);
  const [isAlertOpen, setIsAlertOpen] = useState(true);

  useEffect(() => {
    // Discord連携状態をチェック
    const checkDiscordLink = async () => {
      try {
        const response = await fetch("/api/auth/check-discord-link");
        if (response.ok) {
          const data = await response.json();
          setIsDiscordLinked(data.linked);
        }
      } catch (error) {
        console.error("Discord連携状態の確認に失敗しました:", error);
      }
    };

    checkDiscordLink();
  }, []);

  const handleDiscordLink = () => {
    // Discord OAuth開始
    window.location.href = "/api/oauth/discord?from=settings";
  };

  const handleClose = () => {
    setIsAlertOpen(false);
  };

  // Discord連携済み、または状態をまだ取得していない場合は非表示
  if (isDiscordLinked === null || isDiscordLinked) {
    return null;
  }

  return (
    <Collapse in={isAlertOpen}>
      <Alert
        severity="warning"
        action={
          <>
            <Button
              color="inherit"
              size="small"
              onClick={handleDiscordLink}
              sx={{ mr: 1, fontWeight: "bold" }}
            >
              連携する
            </Button>
            <IconButton
              aria-label="close"
              color="inherit"
              size="small"
              onClick={handleClose}
            >
              <CloseIcon fontSize="inherit" />
            </IconButton>
          </>
        }
        sx={{
          borderRadius: 0,
          "& .MuiAlert-message": {
            width: "100%",
          },
        }}
      >
        <AlertTitle>Discordアカウントの連携が必要です</AlertTitle>
        UniProjectのコミュニティに参加するため、Discordアカウントを連携してください。
        <Typography fontWeight={600}>
          4/1以降に未連携の場合、メンバー登録が抹消される恐れがあります。
        </Typography>
      </Alert>
    </Collapse>
  );
}
