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
  List,
  ListItem,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { useState, useTransition } from "react";
import type { ExternalIdentityData } from "@/classes/ExternalIdentity";
import type { UserData } from "@/classes/types/User";
import { deleteAction } from "./action";

/** Discord brand colour */
const DISCORD_COLOR = "#5865F2";
const DISCORD_HOVER_COLOR = "#4752C4";

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
        p: 3,
        display: "flex",
        flexDirection: "column",
        gap: 3,
      }}
    >
      <Stack spacing={0.5}>
        <Typography variant="h6" sx={{ component: "h3" }}>
          ソーシャルアカウント設定
        </Typography>
        <Typography variant="body2" color="text.secondary">
          外部サービスのアカウントを連携して、ログインに利用できます。
        </Typography>
      </Stack>

      <Stack spacing={2}>
        <Stack direction="row" spacing={2} sx={{ alignItems: "center" }}>
          <Typography variant="subtitle1">Discord</Typography>
          <Divider sx={{ flexGrow: 1 }} />
          {hasDiscordIdentity && (
            <Chip
              label={`連携済み (${discordIdentities.length})`}
              color="success"
              size="small"
              variant="outlined"
            />
          )}
        </Stack>

        {hasDiscordIdentity ? (
          <List
            sx={{ display: "flex", flexDirection: "column", gap: 1.5, p: 0 }}
          >
            {discordIdentities.map((identity) => (
              <ListItem key={identity.id} disablePadding>
                <Box
                  sx={{
                    width: "100%",
                    display: "flex",
                    alignItems: { xs: "flex-start", sm: "center" },
                    flexDirection: { xs: "column", sm: "row" },
                    justifyContent: "space-between",
                    gap: 2,
                    p: 2,
                    borderRadius: 2,
                    bgcolor: "action.hover",
                  }}
                >
                  <Stack
                    direction="row"
                    spacing={2}
                    sx={{
                      alignItems: "center",
                      minWidth: 0,
                    }}
                  >
                    <Avatar
                      src={identity.avatarUrl}
                      alt={resolveDisplayName(identity)}
                      sx={{
                        width: 48,
                        height: 48,
                        bgcolor: DISCORD_COLOR,
                        fontSize: "1.2rem",
                        fontWeight: "bold",
                      }}
                    >
                      {resolveDisplayName(identity).charAt(0).toUpperCase()}
                    </Avatar>

                    <Stack sx={{ minWidth: 0 }}>
                      <Typography
                        variant="subtitle2"
                        sx={{ fontWeight: "bold" }}
                        noWrap
                      >
                        {resolveDisplayName(identity)}
                      </Typography>
                      {identity.username && (
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          noWrap
                        >
                          @{identity.username}
                        </Typography>
                      )}
                      {identity.email && (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          noWrap
                        >
                          {identity.email}
                        </Typography>
                      )}
                    </Stack>
                  </Stack>

                  <Button
                    variant="outlined"
                    color="error"
                    size="small"
                    onClick={() => handleUnlink(identity.id, "Discord")}
                    disabled={isPending}
                    sx={{ alignSelf: { xs: "flex-end", sm: "auto" } }}
                  >
                    連携を解除
                  </Button>
                </Box>
              </ListItem>
            ))}

            <ListItem disablePadding sx={{ mt: 1 }}>
              <Button
                variant="outlined"
                href={`/api/oauth/discord`}
                disabled={isPending}
                sx={{
                  color: DISCORD_COLOR,
                  borderColor: DISCORD_COLOR,
                  "&:hover": {
                    borderColor: DISCORD_HOVER_COLOR,
                    bgcolor: "rgba(88, 101, 242, 0.04)",
                  },
                }}
              >
                別のアカウントを追加
              </Button>
            </ListItem>
          </List>
        ) : (
          <Stack spacing={2} sx={{ alignItems: "flex-start" }}>
            <Typography variant="body2" color="text.secondary">
              Discordアカウントを連携すると、次回からパスワードなしでログインできます。
            </Typography>
            <Button
              variant="contained"
              href={`/api/oauth/discord`}
              disabled={isPending}
              sx={{
                bgcolor: DISCORD_COLOR,
                "&:hover": { bgcolor: DISCORD_HOVER_COLOR },
              }}
              disableElevation
            >
              Discordと連携する
            </Button>
          </Stack>
        )}
      </Stack>

      <UnlinkDialog
        userId={user.id}
        identityId={unlinkingIdentityId}
        provider={unlinkingProvider}
        open={isUnlinkDialogOpen}
        onClose={() => setIsUnlinkDialogOpen(false)}
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
    <SocialAccountsContent
      user={user}
      initialExternalIdentities={externalIdentities}
    />
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
    <Dialog open={open} onClose={() => onClose(false)} maxWidth="xs" fullWidth>
      <DialogTitle>連携解除の確認</DialogTitle>
      <DialogContent>
        <Typography variant="body2">
          本当に {provider} アカウントの連携を解除してもよろしいですか？
        </Typography>
      </DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={() => onClose(false)} color="inherit">
          キャンセル
        </Button>
        <Button
          onClick={handleUnlink}
          color="error"
          variant="contained"
          disableElevation
        >
          連携を解除
        </Button>
      </DialogActions>
    </Dialog>
  );
}
