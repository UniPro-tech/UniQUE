import { Breadcrumbs, Link, Stack, Typography } from "@mui/material";
import RoleCreateForm from "@/components/Forms/RoleCreateForm";

export const metadata = {
  title: "ロール新規作成",
  description: "新しいロールを作成します",
};

export default async function Page() {
  return (
    <Stack spacing={3}>
      <Breadcrumbs aria-label="breadcrumb">
        <Link underline="hover" color="inherit" href="/dashboard/roles">
          ロール管理
        </Link>
        <Typography sx={{ color: "text.primary" }}>新規作成</Typography>
      </Breadcrumbs>

      <Stack>
        <Typography variant="h4" gutterBottom>
          ロール新規作成
        </Typography>
        <Typography variant="body2" color="text.secondary">
          新しいロールを作成し、権限を設定します。
        </Typography>
      </Stack>

      <RoleCreateForm />
    </Stack>
  );
}
