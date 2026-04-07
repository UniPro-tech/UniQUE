import AddIcon from "@mui/icons-material/Add";
import { Box, Button, Stack, Typography } from "@mui/material";
import Link from "next/link";
import ApplicationsDataGrid from "@/components/DataGrids/Applications";

export const metadata = {
  title: "アプリケーション一覧",
  description: "UniQUEのアプリケーション一覧ページ",
};

export default async function Page() {
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
              アプリケーション管理
            </Typography>
            <Typography variant="body2" color="text.secondary">
              サードパーティアプリケーションの一覧を表示・管理します。アプリケーション名をクリックして詳細を編集できます。
            </Typography>
          </Stack>
          <Link
            href="/dashboard/applications/new"
            style={{ textDecoration: "none" }}
          >
            <Button variant="contained" startIcon={<AddIcon />}>
              新規作成
            </Button>
          </Link>
        </Stack>
      </Stack>
      <Box sx={{ width: "100%" }}>
        <ApplicationsDataGrid />
      </Box>
    </Stack>
  );
}
