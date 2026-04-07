"use client";

import { Cancel as CancelIcon, Save as SaveIcon } from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  CircularProgress,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { UserStatus } from "@/classes/types/User";
import {
  type CreateMemberFormData,
  createMember,
  redirectToMemberList,
} from "@/components/Pages/Members/Forms/actions/create";
import {
  getAffiliationPeriodInfo,
  getSelectableAffiliationPeriods,
} from "@/constants/UserConstants";

export default function MemberCreateForm() {
  const router = useRouter();
  const [formData, setFormData] = useState<CreateMemberFormData>({
    customId: "",
    email: "",
    password: "",
    displayName: "",
    status: UserStatus.ACTIVE,
    externalEmail: "",
    affiliationPeriod: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleChange = (field: keyof CreateMemberFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (!formData.customId.trim()) {
      setError("カスタムIDを入力してください");
      setLoading(false);
      return;
    }
    if (!formData.email.trim()) {
      setError("メールアドレスを入力してください");
      setLoading(false);
      return;
    }

    // 内部メールは必ず @uniproject.jp ドメインであること
    if (!formData.email.toLowerCase().endsWith("@uniproject.jp")) {
      setError("内部メールは @uniproject.jp ドメインである必要があります");
      setLoading(false);
      return;
    }

    // 外部メールが入力されている場合は uniproject.jp ドメインであってはならない
    if (formData.externalEmail?.trim()) {
      if (formData.externalEmail.toLowerCase().endsWith("@uniproject.jp")) {
        setError("外部メールは uniproject.jp ドメイン以外を指定してください");
        setLoading(false);
        return;
      }
    }

    try {
      const res = await createMember(formData);
      if (res.success) {
        router.push("/dashboard/members");
      } else {
        setError(res.error || "作成に失敗しました");
      }
    } catch (err) {
      console.error(err);
      setError("予期しないエラーが発生しました");
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async () => {
    await redirectToMemberList();
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        ユーザー新規作成
      </Typography>

      <Box component="form" onSubmit={handleSubmit} noValidate>
        <Stack spacing={3}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          <TextField
            label="カスタムID"
            fullWidth
            required
            value={formData.customId}
            onChange={(e) => handleChange("customId", e.target.value)}
            helperText="ログイン名や識別用のID（英数字）"
            disabled={loading}
          />

          <TextField
            label="メールアドレス"
            fullWidth
            required
            value={formData.email}
            onChange={(e) => handleChange("email", e.target.value)}
            type="email"
            disabled={loading}
          />

          <TextField
            label="一時パスワード"
            fullWidth
            value={formData.password}
            onChange={(e) => handleChange("password", e.target.value)}
            type="password"
            disabled={loading}
            required
          />

          <TextField
            label="表示名"
            fullWidth
            value={formData.displayName}
            onChange={(e) => handleChange("displayName", e.target.value)}
            disabled={loading}
          />

          <TextField
            label="外部メールアドレス (任意)"
            fullWidth
            value={formData.externalEmail}
            onChange={(e) => handleChange("externalEmail", e.target.value)}
            helperText="uniproject.jp 以外のメールアドレスを指定してください（任意）"
            disabled={loading}
            type="email"
          />

          <TextField
            select
            label="所属期 (任意)"
            fullWidth
            value={formData.affiliationPeriod}
            onChange={(e) => handleChange("affiliationPeriod", e.target.value)}
            helperText="所属期を選択（任意）"
            disabled={loading}
          >
            {getSelectableAffiliationPeriods().map((opt) => {
              const info = getAffiliationPeriodInfo(opt.value);
              return (
                <MenuItem key={opt.value} value={opt.value}>
                  <Typography>{opt.value}期</Typography>
                  <Typography
                    variant="caption"
                    sx={{ ml: 1, color: "text.secondary" }}
                  >
                    ({info.calendarYear}年度 {info.label})
                  </Typography>
                </MenuItem>
              );
            })}
          </TextField>

          <TextField
            select
            label="ステータス"
            value={formData.status}
            onChange={(e) => handleChange("status", e.target.value)}
            helperText="ユーザーの状態を選択"
            disabled={loading}
          >
            <MenuItem value={UserStatus.ACTIVE}>active</MenuItem>
            <MenuItem value={UserStatus.ESTABLISHED}>established</MenuItem>
          </TextField>

          <Stack direction="row" spacing={2} justifyContent="flex-end">
            <Button
              variant="outlined"
              startIcon={<CancelIcon />}
              onClick={handleCancel}
              disabled={loading}
            >
              キャンセル
            </Button>

            <Button
              type="submit"
              variant="contained"
              startIcon={
                loading ? <CircularProgress size={20} /> : <SaveIcon />
              }
              disabled={loading}
            >
              {loading ? "作成中..." : "作成"}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Card>
  );
}
