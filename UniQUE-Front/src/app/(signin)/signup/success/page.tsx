import { Stack, Typography } from "@mui/material";

export default async function SignUpSuccessPage({
  searchParams,
}: {
  searchParams: Promise<{ email_verified?: string | string[] }>;
}) {
  const emailVerifiedRaw = (await searchParams).email_verified;
  const emailVerified = Array.isArray(emailVerifiedRaw)
    ? emailVerifiedRaw[0] === "true"
    : emailVerifiedRaw === "true";
  if (!emailVerified) {
    return (
      <Stack>
        <Typography variant="h4" align="center" gutterBottom>
          サインアップが完了しました 🎉
        </Typography>
        <Typography variant="body1" align="center">
          ご登録いただきありがとうございます。ご入力いただいたメールアドレスに確認メールを送信しました。
        </Typography>
        <Typography variant="body1" align="center">
          メール内のリンクをクリックして、メールアドレスの確認を完了してください。
        </Typography>
      </Stack>
    );
  } else {
    return (
      <Stack>
        <Typography variant="h4" align="center" gutterBottom>
          メールアドレスの認証が完了しました 🎉
        </Typography>
        <Typography variant="body1" align="center">
          ご登録いただきありがとうございます。メールアドレスの認証が完了しました。
        </Typography>
        <Typography variant="body1" align="center">
          管理者の承認をお待ちください。承認されると、UniQUEのすべての機能にアクセスできるようになります。
        </Typography>
      </Stack>
    );
  }
}
