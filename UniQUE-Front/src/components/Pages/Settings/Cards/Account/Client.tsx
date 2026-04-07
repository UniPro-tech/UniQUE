"use client";
import ErrorOutlineIcon from "@mui/icons-material/ErrorOutline";
import VerifiedIcon from "@mui/icons-material/Verified";
import {
  Button,
  Chip,
  FormHelperText,
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
  csrfToken,
}: {
  user: UserData;
  csrfToken: string;
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
    <Base
      sid={user.id!}
      action={action}
      isPending={isPending}
      csrfToken={csrfToken}
    >
      <Stack>
        <Typography variant="h5" component={"h3"}>
          基本設定
        </Typography>
        <Typography variant="body2">
          カスタムID、メールアドレスなどの変更を行います。
        </Typography>
      </Stack>
      <TextField
        required
        label="UUID"
        defaultValue={lastResult.user.id}
        disabled
      />
      <input type="hidden" name="id" value={lastResult.user.id} />
      <TextField
        required
        label="表示名"
        defaultValue={lastResult.user.profile?.displayName || ""}
        name="display_name"
      />
      <Stack>
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
      </Stack>
      <Stack>
        <TextField
          required
          label="メールアドレス"
          defaultValue={lastResult.user.email}
          disabled
        />
        <FormHelperText>
          メールアドレスは原則として所属期とカスタムIDに基づいて自動生成されます。
        </FormHelperText>
      </Stack>
      <Stack spacing={1}>
        <TextField
          required
          label="外部メールアドレス"
          defaultValue={
            lastResult.user.pendingEmail || lastResult.user.externalEmail || ""
          }
          name="external_email"
        />
        {(lastResult.user.externalEmail || lastResult.user.pendingEmail) &&
          (lastResult.user.emailVerified && !lastResult.user.pendingEmail ? (
            <Chip
              icon={<VerifiedIcon />}
              label="認証済み"
              color="success"
              size="small"
              sx={{ alignSelf: "flex-start" }}
            />
          ) : (
            <Stack direction="row" spacing={1} alignItems="center">
              <Chip
                icon={<ErrorOutlineIcon />}
                label="未認証"
                color="warning"
                size="small"
              />
              <Button
                variant="outlined"
                size="small"
                onClick={handleResendVerification}
                disabled={isSendingVerification}
              >
                {isSendingVerification ? "送信中..." : "認証メールを送信"}
              </Button>
            </Stack>
          ))}
      </Stack>
      <Stack>
        <TextField
          required
          label="生年月日"
          type="date"
          name="birthdate"
          defaultValue={lastResult.user.profile?.birthdate || ""}
          InputLabelProps={{ shrink: true }}
          disabled={!!lastResult.user.profile?.birthdate}
        />
        <FormHelperText>一度設定したら変更できません。</FormHelperText>
      </Stack>
      <TextField
        label="所属期"
        defaultValue={(lastResult.user.affiliationPeriod || "").toUpperCase()}
        disabled
      />
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
