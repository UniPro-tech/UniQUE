import { Button, Link as MLink, Stack, Typography } from "@mui/material";
import { Session } from "@/classes/Session";
import { User } from "@/classes/User";
import MembersDataGrid from "@/components/DataGrids/Members";
import { PermissionBitsFields } from "@/constants/Permission";

export const metadata = {
  title: "ユーザー一覧",
  description: "UniProjectのメンバー一覧ページ",
};

export default async function Page() {
  const users = await User.getAll();

  // establishedステータスのユーザーを除外
  const filteredUsers = users.filter((u) => u.status !== "established");
  const rows = filteredUsers.map((u) => u.toJson());
  const user = await (await Session.getCurrent())?.getUser();
  const [canDelete, canUpdate, canRead, canCreate] = await Promise.all([
    user?.hasPermission(PermissionBitsFields.USER_DELETE),
    user?.hasPermission(PermissionBitsFields.USER_UPDATE),
    user?.hasPermission(PermissionBitsFields.USER_READ),
    user?.hasPermission(PermissionBitsFields.USER_CREATE),
  ]);

  return (
    <Stack spacing={4}>
      <Stack spacing={2}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
        >
          <Typography variant="h5">ユーザー一覧</Typography>
          <Button
            component="a"
            href="/dashboard/members/new"
            variant="contained"
            disabled={!canCreate}
            sx={{ cursor: canCreate ? "pointer" : "not-allowed" }}
          >
            新規作成
          </Button>
        </Stack>
        <Typography variant="body1">
          UniProjectのメンバー一覧です。
          <br />
          メンバー申請者はここに表示されません。{" "}
          <MLink href="/dashboard/requests">メンバー申請一覧</MLink>
        </Typography>
      </Stack>
      <MembersDataGrid
        rows={rows}
        canDelete={canDelete}
        canUpdate={canUpdate}
        canRead={canRead}
      />
    </Stack>
  );
}
