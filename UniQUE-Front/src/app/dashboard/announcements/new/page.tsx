import { Breadcrumbs, Link, Stack, Typography } from "@mui/material";
import AnnouncementCreateForm from "@/components/Forms/AnnouncementCreateForm";
import { PermissionBitsFields } from "@/constants/Permission";
import { requirePermission } from "@/libs/permissions";

export const metadata = {
  title: "お知らせ新規作成",
  description: "新しいお知らせを作成します",
};

export default async function Page() {
  requirePermission(PermissionBitsFields.ANNOUNCEMENT_CREATE);
  return (
    <Stack spacing={3}>
      <Breadcrumbs aria-label="breadcrumb">
        <Link underline="hover" color="inherit" href="/dashboard/announcements">
          お知らせ管理
        </Link>
        <Typography sx={{ color: "text.primary" }}>新規作成</Typography>
      </Breadcrumbs>

      <Stack>
        <Typography variant="h4" gutterBottom>
          お知らせ新規作成
        </Typography>
        <Typography variant="body2" color="text.secondary">
          公開するお知らせを作成します。
        </Typography>
      </Stack>

      <AnnouncementCreateForm />
    </Stack>
  );
}
