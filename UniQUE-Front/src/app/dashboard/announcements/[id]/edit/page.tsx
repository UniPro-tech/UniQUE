import { Stack, Typography } from "@mui/material";
import { notFound } from "next/navigation";
import { Announcement } from "@/classes/Announcement";
import AnnouncementEditForm from "@/components/Forms/AnnouncementEditForm";
import { PermissionBitsFields } from "@/constants/Permission";
import { requirePermission } from "@/libs/permissions";

export const metadata = {
  title: "お知らせ編集",
  description: "お知らせの編集ページ",
};

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  requirePermission(PermissionBitsFields.ANNOUNCEMENT_UPDATE);
  const { id } = await params;
  const ann = await Announcement.getById(id);
  if (!ann) {
    return notFound();
  }
  const a = ann.toJson();

  return (
    <Stack spacing={3}>
      <Typography variant="h5">お知らせ編集</Typography>
      <AnnouncementEditForm
        id={a.id}
        initial={{ title: a.title, content: a.content, isPinned: !!a.isPinned }}
      />
    </Stack>
  );
}
