import { Stack, Typography } from "@mui/material";

export default async function Authorized() {
  return (
    <Stack>
      <Typography variant="h4" align="center" gutterBottom>
        認可を完了しました ✅
      </Typography>
      <Typography variant="body1" align="center">
        ご利用のデバイス・アプリにお戻りください。
      </Typography>
    </Stack>
  );
}
