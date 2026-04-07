"use client";
import {
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Stack,
  Typography,
} from "@mui/material";
import { SnackbarProvider, useSnackbar } from "notistack";
import { useState, useTransition } from "react";
import type { ExternalIdentityData } from "@/classes/ExternalIdentity";
import type { UserData } from "@/classes/types/User";
import { deleteAction } from "./action";

/** Discord brand colour */
const DISCORD_COLOR = "#5865F2";

/** Resolve a readable name from the identity with fallbacks */
function resolveDisplayName(identity: ExternalIdentityData): string {
  return identity.displayName ?? identity.username ?? identity.externalUserId;
}

function SocialAccountsContent({
  user,
  initialExternalIdentities,
}: {
  user: UserData;
  initialExternalIdentities: ExternalIdentityData[];
}) {
  const [isPending, startTransition] = useTransition();
  const [externalIdentities, setExternalIdentities] = useState<
    ExternalIdentityData[]
  >(initialExternalIdentities);

  // Discordの外部アイデンティティを取得
  const discordIdentities = externalIdentities.filter(
    (identity) => identity.provider.toLowerCase() === "discord",
  );
  const hasDiscordIdentity = discordIdentities.length > 0;

  const [isUnlinkDialogOpen, setIsUnlinkDialogOpen] = useState(false);
  const [unlinkingIdentityId, setUnlinkingIdentityId] = useState("");
  const [unlinkingProvider, setUnlinkingProvider] = useState("");

  const handleUnlink = async (identityId: string, provider: string) => {
    setUnlinkingIdentityId(identityId);
    setUnlinkingProvider(provider);
    setIsUnlinkDialogOpen(true);
  };

  return (
    <Card
      variant="outlined"
      sx={{
        display: "flex",
        p: 2,
        pb: 5,
        flexDirection: "column",
        gap: 2,
      }}
    >
      <Typography variant="h5" component={"h3"}>
        ソーシャルアカウント設定
      </Typography>
      <Typography variant="body2">
        ソーシャルアカウントの連携設定を行います。
      </Typography>

      <Stack direction={"row"} alignItems="center" spacing={1}>
        <Typography variant="h6" textAlign={"left"}>
          Discordアカウント
        </Typography>
        <Divider sx={{ flexGrow: 1 }} />
        {hasDiscordIdentity && (
          <Chip
            label={`連携済み (${discordIdentities.length})`}
            color="success"
            size="small"
          />
        )}
      </Stack>

      {hasDiscordIdentity ? (
        <Stack spacing={2}>
          {discordIdentities.map((identity) => (
            <Box
              key={identity.id}
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 2,
                p: 1.5,
                borderRadius: 2,
                border: "1px solid",
                borderColor: "divider",
                bgcolor: "action.hover",
                flexWrap: "wrap",
              }}
            >
              <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
                <Avatar
                  src={identity.avatarUrl}
                  alt={resolveDisplayName(identity)}
                  sx={{
                    width: 52,
                    height: 52,
                    bgcolor: DISCORD_COLOR,
                    fontSize: "1.3rem",
                  }}
                >
                  {/* Fallback: first letter */}
                  {resolveDisplayName(identity).charAt(0).toUpperCase()}
                </Avatar>

                <Stack sx={{ minWidth: 0 }}>
                  <Typography variant="subtitle1" fontWeight={700} noWrap>
                    {resolveDisplayName(identity)}
                  </Typography>

                  {identity.username && (
                    <Typography variant="body2" color="text.secondary" noWrap>
                      @{identity.username}
                    </Typography>
                  )}

                  {identity.email && (
                    <Typography variant="caption" color="text.secondary" noWrap>
                      {identity.email}
                    </Typography>
                  )}
                </Stack>
              </Box>

              <Button
                variant="outlined"
                color="error"
                onClick={() => handleUnlink(identity.id, "Discord")}
                disabled={isPending}
                sx={{ maxWidth: 300 }}
              >
                連携を解除
              </Button>
            </Box>
          ))}

          <Button
            variant="contained"
            color="primary"
            href={`/api/oauth/discord`}
            disabled={isPending}
            sx={{ maxWidth: 300 }}
          >
            Discordアカウントを追加する
          </Button>
        </Stack>
      ) : (
        <Stack spacing={1}>
          <Typography variant="body2" color="text.secondary">
            Discordアカウントを連携すると、Discord経由でログインができます。
          </Typography>
          <Button
            variant="contained"
            color="primary"
            href={`/api/oauth/discord`}
            disabled={isPending}
            sx={{ maxWidth: 300 }}
          >
            Discordアカウントを連携する
          </Button>
        </Stack>
      )}
      <UnlinkDialog
        userId={user.id}
        identityId={unlinkingIdentityId} // This will be set when opening the dialog
        provider={unlinkingProvider} // This will be set when opening the dialog
        open={isUnlinkDialogOpen} // This will be controlled by state
        onClose={() => setIsUnlinkDialogOpen(false)} // This will be implemented to handle dialog close
        startTransition={startTransition}
        setExternalIdentities={setExternalIdentities}
      />
    </Card>
  );
}

export default function SocialAccountsCardClient({
  user,
  externalIdentities,
}: {
  user: UserData;
  externalIdentities: ExternalIdentityData[];
}) {
  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
      <SocialAccountsContent
        user={user}
        initialExternalIdentities={externalIdentities}
      />
    </SnackbarProvider>
  );
}

function UnlinkDialog({
  userId,
  identityId,
  open,
  onClose,
  provider,
  startTransition,
  setExternalIdentities,
}: {
  userId: string;
  identityId: string;
  open: boolean;
  onClose: (confirmed: boolean) => void;
  provider: string;
  startTransition: (callback: () => void) => void;
  setExternalIdentities: React.Dispatch<
    React.SetStateAction<ExternalIdentityData[]>
  >;
}) {
  const { enqueueSnackbar } = useSnackbar();

  const handleUnlink = async () => {
    startTransition(async () => {
      try {
        await deleteAction(userId, identityId);
        setExternalIdentities((prev) =>
          prev.filter((identity) => identity.id !== identityId),
        );
        enqueueSnackbar(`${provider}アカウントの連携を解除しました`, {
          variant: "success",
        });
        onClose(true);
      } catch (error) {
        console.error("Failed to unlink account:", error);
        enqueueSnackbar(`${provider}アカウントの連携解除に失敗しました`, {
          variant: "error",
        });
      }
    });
  };
  return (
    <Dialog open={open} onClose={() => onClose(false)}>
      <DialogTitle>{`${provider}アカウントの連携を解除しますか?`}</DialogTitle>
      <DialogContent>
        <Typography variant="body2">
          {`${provider}アカウントの連携を解除してもよろしいですか?`}
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={() => onClose(false)} color="primary">
          キャンセル
        </Button>
        <Button onClick={handleUnlink} color="error" variant="contained">
          連携を解除
        </Button>
      </DialogActions>
    </Dialog>
  );
}
