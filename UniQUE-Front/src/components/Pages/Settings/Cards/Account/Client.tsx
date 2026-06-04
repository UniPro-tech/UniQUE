"use client";
import { ErrorOutlineOutlined } from "@mui/icons-material";
import VerifiedIcon from "@mui/icons-material/Verified";
import {
  Box,
  Button,
  Divider,
  FormHelperText,
  InputAdornment,
  Link,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { enqueueSnackbar } from "notistack";
import { useActionState, useEffect, useState } from "react";
import type { UserData } from "@/classes/types/User";
import UserIdChangeApply from "../../Dialogs/UserIdChangeApply";
import Base, { type FormStatus } from "../Base";
import { resendEmailVerificationAction, updateAccountSettings } from "./action";

export default function AccountSettingsCardClient({
  user,
}: {
  user: UserData;
}) {
  const [lastResult, action, isPending] = useActionState(
    updateAccountSettings,
    { user: user, status: null } as {
      user: UserData;
      status: FormStatus | null;
    },
  );
  useEffect(() => {
    if (lastResult.status) {
      enqueueSnackbar(lastResult.status.message, {
        variant: lastResult.status.status,
      });
    }
  }, [lastResult.status]);

  const [openUserIdChangeDialog, setOpenUserIdChangeDialog] = useState(false);
  const [isSendingVerification, setIsSendingVerification] = useState(false);

  const handleResendVerification = async () => {
    setIsSendingVerification(true);
    try {
      const result = await resendEmailVerificationAction(user.id);
      if (result.success) {
        enqueueSnackbar("認証メールを送信しました。メールをご確認ください。", {
          variant: "success",
        });
      } else {
        enqueueSnackbar(result.error || "認証メールの送信に失敗しました。", {
          variant: "error",
        });
      }
    } catch {
      enqueueSnackbar("認証メールの送信に失敗しました。", { variant: "error" });
    } finally {
      setIsSendingVerification(false);
    }
  };
  return (
    <Base sid={user.id} action={action} isPending={isPending}>
      <Stack spacing={2}>
        <Box>
          <Typography variant="h5" component={"h3"}>
            基本設定
          </Typography>
          <Typography variant="body2">
            カスタムID、メールアドレスなどの変更を行います。
          </Typography>
        </Box>

        {/* システム情報（変更不可） */}
        <Box sx={{ p: 2, bgcolor: "background.paper", borderRadius: 1 }}>
          <Typography variant={"subtitle1"} sx={{ mb: 2 }}>
            システム情報
          </Typography>
          <Stack spacing={1}>
            <TextField
              required
              label="UUID"
              defaultValue={lastResult.user.id}
              disabled
            />
            <input type="hidden" name="id" value={lastResult.user.id} />
            <TextField
              required
              label="カスタムID"
              defaultValue={lastResult.user.customId}
              disabled
            />
            <FormHelperText>
              カスタムIDを変更するには申請が必要です。
              <Link href="#" onClick={() => setOpenUserIdChangeDialog(true)}>
                申請する
              </Link>
            </FormHelperText>
            <TextField
              required
              label="メールアドレス"
              defaultValue={lastResult.user.email}
              disabled
            />
            <FormHelperText>
              メールアドレスは原則として所属期とカスタムIDに基づいて自動生成されます。
            </FormHelperText>
            <TextField
              label="所属期"
              defaultValue={(
                lastResult.user.affiliationPeriod || ""
              ).toUpperCase()}
              disabled
            />
          </Stack>
        </Box>

        <Divider />

        {/* ユーザー編集可能情報 */}
        <Box sx={{ p: 2 }}>
          <Typography variant={"subtitle1"} sx={{ mb: 2 }}>
            編集可能情報
          </Typography>
          <Stack spacing={2}>
            <TextField
              required
              label="表示名"
              defaultValue={lastResult.user.profile?.displayName || ""}
              name="display_name"
            />

            <TextField
              required
              label="外部メールアドレス"
              defaultValue={
                lastResult.user.pendingEmail ||
                lastResult.user.externalEmail ||
                ""
              }
              name="external_email"
              slotProps={{
                input: {
                  endAdornment: (
                    <InputAdornment position="end">
                      {lastResult.user.externalEmail ||
                      lastResult.user.pendingEmail ? (
                        lastResult.user.emailVerified &&
                        !lastResult.user.pendingEmail ? (
                          <VerifiedIcon sx={{ color: "success.main" }} />
                        ) : (
                          <Stack
                            direction="row"
                            spacing={1}
                            sx={{ alignItems: "center" }}
                          >
                            <ErrorOutlineOutlined
                              sx={{ color: "warning.main" }}
                            />
                            <Button
                              variant="outlined"
                              size="small"
                              onClick={handleResendVerification}
                              disabled={isSendingVerification}
                            >
                              {isSendingVerification
                                ? "送信中..."
                                : "認証メールを送信"}
                            </Button>
                          </Stack>
                        )
                      ) : null}
                    </InputAdornment>
                  ),
                },
              }}
            />

            <Stack>
              <TextField
                required
                label="生年月日"
                type="date"
                name="birthdate"
                defaultValue={lastResult.user.profile?.birthdate || ""}
                slotProps={{
                  inputLabel: { shrink: true },
                }}
                disabled={!!lastResult.user.profile?.birthdate}
              />
              <FormHelperText>一度設定したら変更できません。</FormHelperText>
            </Stack>
          </Stack>
        </Box>
      </Stack>
      <Button variant="contained" fullWidth type="submit">
        保存
      </Button>
      <UserIdChangeApply
        open={openUserIdChangeDialog}
        handleClose={() => setOpenUserIdChangeDialog(false)}
        user={user}
      />
    </Base>
  );
}
