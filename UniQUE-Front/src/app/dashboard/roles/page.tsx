import AddIcon from "@mui/icons-material/Add";
import { Box, Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import { forbidden } from "next/navigation";
import RolesDataGrid from "@/components/DataGrids/Roles";
import { PermissionBitsFields } from "@/constants/Permission";
import { AuthorizationErrors } from "@/errors/AuthorizationErrors";
import { requirePermission } from "@/libs/permissions";

export const metadata = {
  title: "ロール一覧",
  description: "UniQUEのロール一覧ページ",
};

export default async function Page() {
  try {
    await requirePermission(PermissionBitsFields.ROLE_MANAGE);
  } catch (err) {
    if (err === AuthorizationErrors.AccessDenied) {
      forbidden();
    }
    throw err;
  }

  return (
    <Stack spacing={3}>
      <Stack>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="flex-start"
          spacing={2}
        >
          <Stack>
            <Typography variant="h4" gutterBottom>
              ロール管理
            </Typography>
            <Typography variant="body2" color="text.secondary">
              システムのロール一覧を表示・管理します。ロール名をクリックして詳細を編集できます。
            </Typography>
          </Stack>
          <Link href="/dashboard/roles/new" style={{ textDecoration: "none" }}>
            <Button variant="contained" startIcon={<AddIcon />}>
              新規作成
            </Button>
          </Link>
        </Stack>
      </Stack>
      <Box sx={{ width: "100%" }}>
        <RolesDataGrid />
      </Box>
    </Stack>
  );
}
