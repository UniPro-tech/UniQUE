"use client";

import {
  ArrowBack as ArrowBackIcon,
  Assignment as AssignmentIcon,
  Badge as BadgeIcon,
  Cake as CakeIcon,
  CalendarToday as CalendarIcon,
  Edit as EditIcon,
  Email as EmailIcon,
  Link as LinkIcon,
  Password,
  Person as PersonIcon,
  Web as WebIcon,
  X as XIcon,
} from "@mui/icons-material";
import {
  Avatar,
  Box,
  Button,
  Card,
  Chip,
  Divider,
  Grid,
  Link,
  Stack,
  Typography,
  useTheme,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { SnackbarProvider } from "notistack";
import { type ReactNode, useState } from "react";
import type { ExternalIdentityData } from "@/classes/ExternalIdentity";
import { type UserData, UserStatus } from "@/classes/types/User";
import PasswordResetAdmin from "@/components/Dialogs/PasswordResetAdmin";
import ProfileEditForm from "@/components/Forms/ProfileEditForm";
import {
  getAffiliationPeriodLabel,
  getStatusLabel,
} from "@/constants/UserConstants";

interface ProfileProps {
  user: UserData;
  externalIdentities: ExternalIdentityData[];
  variant?: "self" | "detail" | "admin";
  onBack?: () => void;
  showTimestamps?: boolean;
}

// プロフィール情報の1項目を表示するためのヘルパーコンポーネント
const InfoItem = ({
  icon,
  label,
  children,
}: {
  icon: ReactNode;
  label: string;
  children: ReactNode;
}) => (
  <Box sx={{ display: "flex", alignItems: "flex-start", gap: 1.5 }}>
    <Box sx={{ color: "text.secondary", mt: 0.5 }}>{icon}</Box>
    <Box>
      <Typography
        variant="caption"
        color="text.secondary"
        sx={{ display: "block" }}
      >
        {label}
      </Typography>
      <Box sx={{ mt: 0.5 }}>{children}</Box>
    </Box>
  </Box>
);

// プロバイダ名を綺麗に表示するためのヘルパー
const getProviderLabel = (provider: string) => {
  switch (provider.toLowerCase()) {
    case "google.com":
      return "Google";
    case "github.com":
      return "GitHub";
    case "twitter.com":
    case "x.com":
      return "X (Twitter)";
    case "discord":
      return "Discord";
    default:
      return provider;
  }
};

export default function ProfileClient({
  user,
  externalIdentities = [],
  variant = "self",
  onBack,
  showTimestamps = false,
}: ProfileProps) {
  const router = useRouter();
  const theme = useTheme();

  const [userProfile, setUserProfile] = useState(user.profile);
  const hasFullData = user?.email !== undefined || user?.status !== undefined;
  const [editMode, setEditMode] = useState(false);
  const [passwordResetOpen, setPasswordResetOpen] = useState(false);

  const displayName =
    userProfile?.displayName || user?.customId || "名称未設定";

  return (
    <Stack spacing={4} sx={{ maxWidth: 900, mx: "auto", width: "100%" }}>
      {/* ページヘッダー部分 */}
      <Stack
        direction={{ xs: "column", sm: "row" }}
        sx={{
          justifyContent: "space-between",
          alignItems: { xs: "flex-start", sm: "center" },
        }}
        spacing={2}
      >
        <Box>
          {(variant === "detail" || variant === "admin") && (
            <Button
              startIcon={<ArrowBackIcon />}
              onClick={() => {
                if (onBack) onBack();
                else if (window.history.length > 2) router.back();
                else router.push("/dashboard/members");
              }}
              sx={{ mb: 1, ml: -1 }}
              color="inherit"
            >
              メンバー一覧に戻る
            </Button>
          )}
          <Typography variant="h4" gutterBottom sx={{ fontWeight: "bold" }}>
            {variant === "self" ? "プロフィール" : "ユーザー詳細"}
          </Typography>
          {!hasFullData && (
            <Chip
              label="公開情報のみ表示中"
              size="small"
              color="warning"
              variant="outlined"
              sx={{ mt: 1 }}
            />
          )}
        </Box>

        <Stack
          direction="row"
          spacing={2}
          sx={{ flexWrap: "wrap" }}
          useFlexGap={true}
        >
          {(variant === "self" || variant === "admin") && !editMode && (
            <Button
              variant="contained"
              startIcon={<EditIcon />}
              onClick={() => setEditMode(true)}
            >
              プロフ編集
            </Button>
          )}
          {variant === "admin" && (
            <Button
              variant="outlined"
              color="error"
              startIcon={<Password />}
              onClick={() => setPasswordResetOpen(true)}
            >
              パスワードリセット
            </Button>
          )}
        </Stack>
      </Stack>

      {/* メインカード（基本情報） */}
      <Card
        sx={{
          p: { xs: 3, sm: 4 },
          borderRadius: 3,
          boxShadow: theme.shadows[2],
        }}
      >
        <Stack spacing={4}>
          {/* ユーザー基本情報 (アバターと名前) */}
          <Box sx={{ display: "flex", gap: 3, alignItems: "center" }}>
            <Avatar
              sx={{
                width: 80,
                height: 80,
                fontSize: "2rem",
                bgcolor: theme.palette.primary.main,
              }}
            >
              {displayName.charAt(0).toUpperCase()}
            </Avatar>
            <Box>
              <Typography variant="h5" sx={{ fontWeight: "bold" }} gutterBottom>
                {displayName}
              </Typography>
              <Stack
                direction="row"
                spacing={1}
                useFlexGap={true}
                sx={{ flexWrap: "wrap" }}
              >
                {user?.status && (
                  <Chip
                    label={
                      typeof user.status === "string"
                        ? user.status
                        : getStatusLabel(user.status as UserStatus)
                    }
                    size="small"
                    color={
                      user?.status === UserStatus.ACTIVE
                        ? "success"
                        : user?.status === UserStatus.SUSPENDED
                          ? "error"
                          : user?.status === UserStatus.ARCHIVED
                            ? "default"
                            : "primary"
                    }
                  />
                )}
                {user?.affiliationPeriod && (
                  <Chip
                    label={`${getAffiliationPeriodLabel(
                      user.affiliationPeriod as unknown as string | null,
                    )}期`}
                    color="primary"
                    variant="outlined"
                    size="small"
                  />
                )}
                {user?.emailVerified && (
                  <Chip
                    label="メール認証済"
                    size="small"
                    color="success"
                    variant="outlined"
                  />
                )}
                {userProfile && userProfile.isAdult === false && (
                  <Chip label="U-18" size="small" color="warning" />
                )}
              </Stack>
            </Box>
          </Box>

          <Divider />

          {/* 編集モード / 閲覧モード */}
          {editMode ? (
            <SnackbarProvider maxSnack={3} autoHideDuration={6000}>
              <ProfileEditForm
                userId={user.id}
                profile={userProfile}
                onCancel={() => setEditMode(false)}
                onSuccess={() => setEditMode(false)}
                setProfile={setUserProfile}
              />
            </SnackbarProvider>
          ) : (
            <Box>
              <Typography
                variant="subtitle2"
                color="text.secondary"
                sx={{ mb: 3 }}
              >
                詳細情報
              </Typography>
              <Grid container spacing={4}>
                <Grid size={{ xs: 12, sm: 6 }}>
                  <InfoItem icon={<PersonIcon />} label="表示名">
                    <Typography variant="body1" sx={{ fontWeight: 500 }}>
                      {userProfile?.displayName || "未設定"}
                    </Typography>
                  </InfoItem>
                </Grid>

                <Grid size={{ xs: 12, sm: 6 }}>
                  <InfoItem icon={<BadgeIcon />} label="カスタムID">
                    <Typography variant="body1" sx={{ fontWeight: 500 }}>
                      {user?.customId || "未設定"}
                    </Typography>
                  </InfoItem>
                </Grid>

                {user?.email && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<EmailIcon />} label="メールアドレス">
                      <Typography variant="body1">{user.email}</Typography>
                    </InfoItem>
                  </Grid>
                )}

                {user?.externalEmail && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<EmailIcon />} label="外部メールアドレス">
                      <Typography variant="body1">
                        {user.externalEmail}
                      </Typography>
                    </InfoItem>
                  </Grid>
                )}

                {userProfile?.websiteUrl && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<WebIcon />} label="ウェブサイト">
                      <Link
                        href={userProfile.websiteUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        underline="hover"
                      >
                        {userProfile.websiteUrl}
                      </Link>
                    </InfoItem>
                  </Grid>
                )}

                {userProfile?.twitterHandle && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<XIcon />} label="Twitter (X)">
                      <Link
                        href={`https://twitter.com/${userProfile.twitterHandle}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        underline="hover"
                      >
                        @{userProfile.twitterHandle}
                      </Link>
                    </InfoItem>
                  </Grid>
                )}

                {userProfile?.birthdate && userProfile?.birthdateVisible && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<CakeIcon />} label="誕生日">
                      <Typography variant="body1">
                        {new Date(userProfile.birthdate).toLocaleDateString(
                          "ja-JP",
                          {
                            year: "numeric",
                            month: "long",
                            day: "numeric",
                          },
                        )}
                      </Typography>
                    </InfoItem>
                  </Grid>
                )}

                {userProfile?.joinedAt && (
                  <Grid size={{ xs: 12, sm: 6 }}>
                    <InfoItem icon={<CalendarIcon />} label="参加日">
                      <Typography variant="body1">
                        {new Date(userProfile.joinedAt).toLocaleDateString(
                          "ja-JP",
                          {
                            year: "numeric",
                            month: "long",
                            day: "numeric",
                          },
                        )}
                      </Typography>
                    </InfoItem>
                  </Grid>
                )}

                {userProfile?.bio && (
                  <Grid size={{ xs: 12 }} sx={{ mt: 1 }}>
                    <InfoItem icon={<AssignmentIcon />} label="自己紹介">
                      <Typography
                        variant="body1"
                        sx={{ whiteSpace: "pre-wrap", lineHeight: 1.7 }}
                      >
                        {userProfile.bio}
                      </Typography>
                    </InfoItem>
                  </Grid>
                )}
              </Grid>
            </Box>
          )}
        </Stack>
      </Card>

      {/* 連携アカウント一覧カード */}
      <Card
        sx={{
          p: { xs: 3, sm: 4 },
          borderRadius: 3,
          boxShadow: theme.shadows[2],
        }}
      >
        <Typography
          variant="h6"
          sx={{
            mb: 3,
            display: "flex",
            alignItems: "center",
            fontWeight: "bold",
            gap: 1.5,
          }}
        >
          <LinkIcon color="primary" />
          連携アカウント一覧
        </Typography>

        {externalIdentities.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            連携済みの外部アカウントはありません。
          </Typography>
        ) : (
          <Grid container spacing={2}>
            {externalIdentities.map((identity) => (
              <Grid size={{ xs: 12, sm: 6 }} key={identity.id}>
                <Box
                  sx={{
                    p: 2,
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 2,
                    display: "flex",
                    alignItems: "center",
                    gap: 2,
                    bgcolor: "background.paper",
                    transition: "box-shadow 0.2s",
                    "&:hover": {
                      boxShadow: theme.shadows[1],
                    },
                  }}
                >
                  <Avatar
                    src={identity.avatarUrl}
                    alt={identity.displayName || identity.username}
                    sx={{ width: 44, height: 44 }}
                  >
                    {/* 万が一画像がない場合のフォールバック */}
                    {(identity.displayName || identity.provider)
                      .charAt(0)
                      .toUpperCase()}
                  </Avatar>
                  <Box sx={{ flexGrow: 1, minWidth: 0 }}>
                    <Stack
                      direction="row"
                      spacing={1}
                      sx={{ alignItems: "center", mb: 0.5 }}
                    >
                      <Typography
                        variant="subtitle2"
                        noWrap
                        sx={{ maxWidth: "100%", fontWeight: "bold" }}
                      >
                        {identity.displayName || identity.username || "未設定"}
                      </Typography>
                      <Chip
                        label={getProviderLabel(identity.provider)}
                        size="small"
                        variant="outlined"
                        color="primary"
                        sx={{
                          height: 20,
                          fontSize: "0.7rem",
                          fontWeight: "500",
                        }}
                      />
                    </Stack>
                    {identity.email && (
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        noWrap
                        sx={{ display: "block" }}
                      >
                        {identity.email}
                      </Typography>
                    )}
                    {/* preffered username を表示 */}
                    {identity.username && identity.displayName && (
                      <Typography
                        variant="caption"
                        color="text.disabled"
                        noWrap
                        sx={{
                          display: "block",
                          fontFamily: "monospace",
                        }}
                      >
                        Username: {identity.username}
                      </Typography>
                    )}
                    {/* 管理者画面の時はプロバイダ側のsubを表示 */}
                    {variant === "admin" && (
                      <Typography
                        variant="caption"
                        color="text.disabled"
                        noWrap
                        sx={{
                          display: "block",
                          fontFamily: "monospace",
                        }}
                      >
                        Ext-ID: {identity.externalUserId}
                      </Typography>
                    )}
                  </Box>
                </Box>
              </Grid>
            ))}
          </Grid>
        )}
      </Card>

      {/* フッター (タイムスタンプ) */}
      {showTimestamps && (user?.createdAt || user?.updatedAt) && (
        <Stack direction="row" spacing={3} sx={{ justifyContent: "flex-end" }}>
          {user?.createdAt && (
            <Typography variant="caption" color="text.disabled">
              作成:{" "}
              {new Date(user.createdAt).toLocaleDateString("ja-JP", {
                year: "numeric",
                month: "short",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </Typography>
          )}
          {user?.updatedAt && (
            <Typography variant="caption" color="text.disabled">
              更新:{" "}
              {new Date(user.updatedAt).toLocaleDateString("ja-JP", {
                year: "numeric",
                month: "short",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </Typography>
          )}
        </Stack>
      )}

      <SnackbarProvider>
        <PasswordResetAdmin
          open={passwordResetOpen}
          userId={user.id}
          onClose={() => setPasswordResetOpen(false)}
        />
      </SnackbarProvider>
    </Stack>
  );
}
