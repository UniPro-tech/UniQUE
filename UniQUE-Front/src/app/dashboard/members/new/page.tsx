import { Breadcrumbs, Link, Stack, Typography } from "@mui/material";
import MemberCreateForm from "@/components/Pages/Members/Forms/MemberCreateForm";
import { PermissionBitsFields } from "@/constants/Permission";
import { requirePermission } from "@/libs/permissions";

export const metadata = {
  title: "ユーザー手動作成",
  description: "管理者が手動で新しいユーザーを作成します",
};

export default async function Page() {
  requirePermission(PermissionBitsFields.USER_CREATE);
  return (
    <Stack spacing={3}>
      <Breadcrumbs aria-label="breadcrumb">
        <Link underline="hover" color="inherit" href="/dashboard/members">
          ユーザー管理
        </Link>
        <Typography sx={{ color: "text.primary" }}>新規作成</Typography>
      </Breadcrumbs>

      <Stack>
        <Typography variant="h4" gutterBottom>
          ユーザー手動作成
        </Typography>
        <Typography variant="body2" color="text.secondary">
          管理者権限で新規ユーザーを手動で作成します。
        </Typography>
      </Stack>

      <MemberCreateForm />
    </Stack>
  );
}
