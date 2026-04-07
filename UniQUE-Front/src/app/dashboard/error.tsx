"use client";

import ErrorIcon from "@mui/icons-material/Error";
import HomeIcon from "@mui/icons-material/Home";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import {
  Alert,
  Box,
  Button,
  Card,
  Container,
  Stack,
  Typography,
} from "@mui/material";
import Link from "next/link";

interface ErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

/**
 * ダッシュボード用のエラーバウンダリ
 * アクセス拒否エラーと一般エラーを区別して表示
 */
export default function DashboardError({ error, reset }: ErrorProps) {
  // アクセス拒否エラー用の表示
  if (error.message.includes("B0001")) {
    return (
      <Container maxWidth="sm">
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
            minHeight: "100vh",
            py: 4,
          }}
        >
          <Card
            variant="outlined"
            sx={{
              width: "100%",
              p: 4,
              textAlign: "center",
              borderRadius: 2,
            }}
          >
            <Box
              sx={{
                display: "flex",
                justifyContent: "center",
                mb: 2,
              }}
            >
              <ErrorIcon
                sx={{
                  fontSize: 64,
                  color: "error.main",
                }}
              />
            </Box>

            <Typography
              variant="h5"
              component="h1"
              gutterBottom
              sx={{ fontWeight: 600 }}
            >
              アクセス権限がありません
            </Typography>

            <Typography variant="body2" color="textSecondary" sx={{ my: 2 }}>
              このページへのアクセス権限がないようです。
            </Typography>

            <Alert severity="info" sx={{ my: 3, textAlign: "left" }}>
              <Typography variant="caption">
                <strong>権限が必要な場合：</strong>
              </Typography>
              <Typography variant="caption" component="div" sx={{ mt: 1 }}>
                プロジェクト管理者に権限の付与をリクエストしてください。
              </Typography>
            </Alert>

            <Stack direction="row" spacing={2} sx={{ mt: 4 }}>
              <Button
                component={Link}
                href="/dashboard"
                variant="contained"
                fullWidth
                startIcon={<HomeIcon />}
              >
                ダッシュボードに戻る
              </Button>
            </Stack>
          </Card>
        </Box>
      </Container>
    );
  }

  // 一般エラー用の表示
  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          minHeight: "100vh",
          py: 4,
        }}
      >
        <Card
          variant="outlined"
          sx={{
            width: "100%",
            p: 4,
            textAlign: "center",
            borderRadius: 2,
          }}
        >
          <Box
            sx={{
              display: "flex",
              justifyContent: "center",
              mb: 2,
            }}
          >
            <ErrorIcon
              sx={{
                fontSize: 64,
                color: "warning.main",
              }}
            />
          </Box>

          <Typography
            variant="h5"
            component="h1"
            gutterBottom
            sx={{ fontWeight: 600 }}
          >
            エラーが発生しました
          </Typography>

          <Typography variant="body2" color="textSecondary" sx={{ my: 2 }}>
            予期しないエラーが発生しました。
          </Typography>

          {process.env.NODE_ENV === "development" && (
            <Alert severity="warning" sx={{ my: 3, textAlign: "left" }}>
              <Typography
                variant="caption"
                component="div"
                sx={{ wordBreak: "break-word" }}
              >
                <strong>エラー詳細（開発環境）：</strong>
              </Typography>
              <Typography
                variant="caption"
                component="div"
                sx={{
                  mt: 1,
                  p: 1,
                  backgroundColor: "grey.100",
                  borderRadius: 1,
                  fontFamily: "monospace",
                  fontSize: "0.75rem",
                  overflow: "auto",
                  maxHeight: 200,
                }}
              >
                {error.message}
              </Typography>
            </Alert>
          )}

          <Stack direction="row" spacing={2} sx={{ mt: 4 }}>
            <Button
              variant="outlined"
              fullWidth
              startIcon={<RestartAltIcon />}
              onClick={() => reset()}
            >
              もう一度試す
            </Button>
            <Button
              component={Link}
              href="/dashboard"
              variant="contained"
              fullWidth
              startIcon={<HomeIcon />}
            >
              ダッシュボードに戻る
            </Button>
          </Stack>
        </Card>
      </Box>
    </Container>
  );
}
