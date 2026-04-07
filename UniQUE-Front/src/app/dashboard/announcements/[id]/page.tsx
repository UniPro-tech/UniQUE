import { Box, Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import { notFound } from "next/navigation";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Announcement } from "@/classes/Announcement";
import { Session } from "@/classes/Session";
import TemporarySnackProvider, {
  type SnackbarData,
} from "@/components/TemporarySnackProvider";
import { PermissionBitsFields } from "@/constants/Permission";

export const metadata = {
  title: "アナウンス詳細",
  description: "アナウンスの詳細表示ページ",
};

export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ success?: string; type?: string; error?: string }>;
}) {
  const { id } = await params;
  const ann = await Announcement.getById(id);

  if (!ann) {
    notFound();
  }

  const a = ann.toJson();

  const { success, type, error } = await searchParams;

  const snacks: SnackbarData[] = [];
  if (success) {
    snacks.push({
      message:
        type === "delete"
          ? "お知らせを削除しました"
          : type === "edit"
            ? "お知らせを更新しました"
            : "操作に成功しました",
      variant: "success",
    });
  } else if (error) {
    snacks.push({ message: `エラー: ${error}`, variant: "error" });
  }

  const user = await (await Session.getCurrent())?.getUser();
  const canEdit = await user?.hasPermission(
    PermissionBitsFields.ANNOUNCEMENT_UPDATE,
  );

  return (
    <Stack spacing={3}>
      <TemporarySnackProvider snacks={snacks} />
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Typography variant="h5">お知らせ詳細</Typography>
        <Box>
          <Link
            href="/dashboard/announcements"
            style={{ textDecoration: "none", marginRight: 8 }}
          >
            <Button variant="outlined">一覧へ戻る</Button>
          </Link>
          {canEdit && (
            <Link
              href={`/dashboard/announcements/${a.id}/edit`}
              style={{ textDecoration: "none", marginLeft: 8 }}
            >
              <Button variant="contained">編集</Button>
            </Link>
          )}
        </Box>
      </Stack>

      <Box>
        <Typography variant="h5">{a.title}</Typography>
        <Typography variant="caption" color="text.secondary">
          作成者:{" "}
          {a.createdBy?.profile?.displayName ||
            a.createdBy?.customId ||
            a.createdBy?.id ||
            "system"}{" "}
          • {new Date(a.createdAt).toLocaleString()}
        </Typography>
        <Box mt={2} style={{ whiteSpace: "pre-wrap" }}>
          <div className="markdown">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {a.content}
            </ReactMarkdown>
          </div>
        </Box>
      </Box>
    </Stack>
  );
}
