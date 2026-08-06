"use client";

import { TextField } from "@mui/material";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import MuiCard from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import Divider from "@mui/material/Divider";
import Stack from "@mui/material/Stack";
import type { Theme } from "@mui/material/styles";
import Typography from "@mui/material/Typography";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { getAuthRequest } from "./action";

export default function DeviceFlowUserCodeInput() {
  const [code, setCode] = useState<string[]>(["", "", "", "", "", "", "", ""]);
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);
  const router = useRouter();

  const handleChange = (value: string, index: number) => {
    // 英数のみ許可
    if (value && !/^[a-zA-Z0-9]$/.test(value)) return;

    const newCode = [...code];
    newCode[index] = value;
    setCode(newCode);

    // 文字が入力されたら、次の入力欄に自動フォーカス
    if (value && index < 7) {
      inputRefs.current[index + 1]?.focus();
    }
  };
  // キーボード操作（BackSpaceキーで戻る処理）の制御
  const handleKeyDown = (
    e: React.KeyboardEvent<HTMLDivElement>,
    index: number,
  ) => {
    // 空の状態でBackSpaceが押されたら、前の入力欄に自動フォーカス
    if (e.key === "Backspace" && !code[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };
  // ペースト（貼り付け）時の処理
  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault(); // 通常のペースト動作をキャンセル
    const pastedData = e.clipboardData.getData("text").trim();

    // 英数のみ、かつ4桁のデータかチェック
    if (/^[a-zA-Z0-9]$/.test(pastedData)) {
      const newCode = pastedData.split("");
      setCode(newCode);
      // 一番最後の入力欄にフォーカスを移動
      inputRefs.current[7]?.focus();
    }
  };
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
        <Box>
          <Typography component="h1" variant="h6">
            デバイス認証コードの入力
          </Typography>
          <Typography variant="caption" color="text.secondary">
            デバイスに表示されている認証コードを入力してください。
          </Typography>
        </Box>
      </Box>

      <Box sx={{ mt: 1 }}>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          信頼できるデバイスに表示されている認証コードを入力してください。
          <br />
          認証コードは、デバイスの画面に表示される8桁の英数字です。
        </Typography>
      </Box>

      <Divider sx={{ my: 1.5 }} />

      <CardContent sx={{ px: 0 }}>
        <Stack sx={{ flexDirection: "row", gap: 2 }}>
          {code.map((digit, index) => (
            <TextField
              // biome-ignore lint/suspicious/noArrayIndexKey: index以外のユニークな値がないため、indexをkeyとして使用
              key={index}
              value={digit}
              onChange={(e) => handleChange(e.target.value, index)}
              onKeyDown={(e) => handleKeyDown(e, index)}
              onPaste={handlePaste}
              slotProps={{
                input: {
                  inputProps: {
                    maxLength: 1,
                    inputMode: "numeric",
                  },
                },
              }}
              sx={{
                "& input": {
                  textAlign: "center",
                  fontSize: "1.5rem",
                  width: "40px",
                },
              }}
              inputRef={(el: HTMLInputElement | null) => {
                inputRefs.current[index] = el;
              }}
              variant="outlined"
            />
          ))}
        </Stack>
      </CardContent>

      <CardActions sx={{ px: 0, mt: 1 }}>
        <Box sx={{ display: "flex", gap: 2, width: "100%", flexWrap: "wrap" }}>
          <Box
            component="form"
            action={async () => {
              const authReqID = await getAuthRequest(code.join());
              router.push(`/device/consent/${authReqID}`);
            }}
            method="post"
            sx={{ flex: 1 }}
          >
            <Button type="submit" variant="contained" color="primary" fullWidth>
              進む
            </Button>
          </Box>
        </Box>
      </CardActions>
    </MuiCard>
  );
}
