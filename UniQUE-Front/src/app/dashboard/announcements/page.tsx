import AddIcon from "@mui/icons-material/Add";
import { Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import { Announcement } from "@/classes/Announcement";
import { Session } from "@/classes/Session";
import AnnouncementsList from "@/components/Lists/AnnouncementsList";
import { PermissionBitsFields } from "@/constants/Permission";

export const metadata = {
  title: "アナウンス一覧",
  description: "アナウンスの一覧です。",
};

export default async function Page() {
  const anns = (await Announcement.getAll()).map((a) => a.toJson());

  const user = await (await Session.getCurrent())?.getUser();

  const [canCreate, canUpdate, canDelete, canPin] = await Promise.all([
    user?.hasPermission(PermissionBitsFields.ANNOUNCEMENT_CREATE),
    user?.hasPermission(PermissionBitsFields.ANNOUNCEMENT_UPDATE),
    user?.hasPermission(PermissionBitsFields.ANNOUNCEMENT_DELETE),
    user?.hasPermission(PermissionBitsFields.ANNOUNCEMENT_PIN),
  ]);

  return (
    <Stack spacing={3}>
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Stack>
          <Typography variant="h5">アナウンス一覧</Typography>
          <Typography variant="body2" color="text.secondary">
            アナウンスの一覧です。
          </Typography>
        </Stack>
        {canCreate && (
          <Link
            href="/dashboard/announcements/new"
            style={{ textDecoration: "none" }}
          >
            <Button variant="contained" startIcon={<AddIcon />}>
              新規作成
            </Button>
          </Link>
        )}
      </Stack>

      <AnnouncementsList
        initial={anns}
        canUpdate={canUpdate}
        canDelete={canDelete}
        canPin={canPin}
      />
    </Stack>
  );
}
