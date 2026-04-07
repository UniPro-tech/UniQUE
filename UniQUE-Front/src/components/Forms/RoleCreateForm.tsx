"use client";

import { Cancel as CancelIcon, Save as SaveIcon } from "@mui/icons-material";
import {
  Alert,
  Box,
  Button,
  Card,
  Checkbox,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Divider,
  FormControlLabel,
  FormGroup,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";
import {
  type CreateRoleFormData,
  createRole,
  redirectToRoleList,
} from "@/app/dashboard/roles/new/action";
import { PermissionBitsFields, PermissionTexts } from "@/constants/Permission";
import { PermissionGroups } from "./PermissionGroups";

interface RoleCreateFormProps {
  onCancel?: () => void;
}

export default function RoleCreateForm({ onCancel }: RoleCreateFormProps) {
  const router = useRouter();
  const [formData, setFormData] = useState<CreateRoleFormData>({
    customId: "",
    name: "",
    description: "",
    permissionBitmask: 0n,
    isDefault: false,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleChange = (field: keyof CreateRoleFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setError(null);
  };

  const handlePermissionToggle = (permission: bigint) => {
    setFormData((prev) => ({
      ...prev,
      permissionBitmask: prev.permissionBitmask ^ permission,
    }));
  };

  const isPermissionChecked = (permission: bigint) => {
    return (formData.permissionBitmask & permission) !== 0n;
  };

  const getPermissionName = (permission: bigint): string => {
    const key = Object.keys(PermissionBitsFields).find(
      (k) =>
        PermissionBitsFields[k as keyof typeof PermissionBitsFields] ===
        permission,
    );
    return key
      ? PermissionTexts[key as keyof typeof PermissionTexts]
      : "不明な権限";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // バリデーション
    if (!formData.customId.trim()) {
      setError("カスタムIDを入力してください");
      setLoading(false);
      return;
    }

    if (!formData.name.trim()) {
      setError("ロール名を入力してください");
      setLoading(false);
      return;
    }

    // If isDefault is checked, ask whether to assign to existing users
    if (formData.isDefault) {
      setConfirmOpen(true);
      setLoading(false);
      return;
    }

    try {
      const result = await createRole(formData);

      if (result.success) {
        // 成功したらロール一覧ページにリダイレクト
        router.push("/dashboard/roles");
      } else {
        setError(result.error || "作成に失敗しました");
      }
    } catch (err) {
      setError("予期しないエラーが発生しました");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleIsDefaultToggle = (v: boolean) => {
    setFormData((prev) => ({ ...prev, isDefault: v }));
    setError(null);
  };

  const [confirmOpen, setConfirmOpen] = useState(false);

  const handleConfirm = async (assignExisting: boolean) => {
    setConfirmOpen(false);
    setLoading(true);
    try {
      const result = await createRole({ ...formData, assignExisting });
      if (result.success) {
        router.push("/dashboard/roles");
      } else {
        setError(result.error || "作成に失敗しました");
      }
    } catch (err) {
      console.error(err);
      setError("予期しないエラーが発生しました");
    } finally {
      setLoading(false);
    }
  };

  const handleCancelClick = async () => {
    if (onCancel) {
      onCancel();
    } else {
      await redirectToRoleList();
    }
  };

  return (
    <Card sx={{ p: 3 }}>
      <Typography variant="h6" gutterBottom>
        ロール新規作成
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
            helperText="ロールを識別するためのユニークなIDを入力してください（例: admin, member）"
            disabled={loading}
          />

          <TextField
            label="ロール名"
            fullWidth
            required
            value={formData.name}
            onChange={(e) => handleChange("name", e.target.value)}
            helperText="ロールの表示名を入力してください"
            disabled={loading}
          />

          <TextField
            label="説明"
            fullWidth
            multiline
            rows={3}
            value={formData.description}
            onChange={(e) => handleChange("description", e.target.value)}
            helperText="ロールの説明を入力してください（任意）"
            disabled={loading}
          />

          <Divider />

          <FormControlLabel
            control={
              <Checkbox
                checked={!!formData.isDefault}
                onChange={(e) => handleIsDefaultToggle(e.target.checked)}
                disabled={loading}
              />
            }
            label="新規ユーザーにデフォルトで付与する"
          />

          <Dialog
            open={confirmOpen}
            onClose={() => setConfirmOpen(false)}
            aria-labelledby="confirm-assign-existing-title"
          >
            <DialogTitle id="confirm-assign-existing-title">
              既存ユーザーへの付与
            </DialogTitle>
            <DialogContent>
              <DialogContentText>
                新規ユーザーにデフォルトで付与する設定にしています。
                既存ユーザー（status が active または established
                のユーザー）にも このロールを付与しますか？
              </DialogContentText>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => handleConfirm(false)} disabled={loading}>
                付与しない
              </Button>
              <Button
                onClick={() => handleConfirm(true)}
                variant="contained"
                disabled={loading}
              >
                全員に付与する
              </Button>
            </DialogActions>
          </Dialog>

          <Box>
            <Typography variant="subtitle1" gutterBottom>
              権限設定
            </Typography>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              このロールに付与する権限を選択してください
            </Typography>

            {PermissionGroups.map((group) => (
              <Box key={group.title} sx={{ mb: 2 }}>
                <Typography
                  variant="subtitle2"
                  color="primary"
                  sx={{ mt: 2, mb: 1 }}
                >
                  {group.title}
                </Typography>
                <FormGroup>
                  {group.permissions.map((permission) => (
                    <FormControlLabel
                      key={permission}
                      control={
                        <Checkbox
                          checked={isPermissionChecked(permission)}
                          onChange={() => handlePermissionToggle(permission)}
                          disabled={loading}
                        />
                      }
                      label={getPermissionName(permission)}
                    />
                  ))}
                </FormGroup>
              </Box>
            ))}

            <Typography variant="caption" color="text.secondary" sx={{ mt: 2 }}>
              選択された権限ビットマスク: {formData.permissionBitmask}
            </Typography>
          </Box>

          <Stack direction="row" spacing={2} justifyContent="flex-end">
            <Button
              variant="outlined"
              startIcon={<CancelIcon />}
              onClick={handleCancelClick}
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
