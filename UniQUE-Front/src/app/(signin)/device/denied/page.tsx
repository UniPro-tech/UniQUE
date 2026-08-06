import { Stack, Typography } from "@mui/material";

export default async function Denied() {
  return (
    <Stack>
      <Typography variant="h4" align="center" gutterBottom>
        認可を却下しました ❌
      </Typography>
      <Typography variant="body1" align="center">
        ご利用のアプリ・デバイスにお戻りください。
      </Typography>
    </Stack>
  );
}
