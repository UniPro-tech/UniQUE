import { Box, Breadcrumbs, Link, Stack, Typography } from "@mui/material";
import { notFound } from "next/navigation";
import { Suspense } from "react";
import { Role } from "@/classes/Role";
import RoleUsersDataGrid from "@/components/DataGrids/RoleUsers";
import RoleEditForm from "@/components/Forms/RoleEditForm";
import RoleBulkAssignButton from "@/components/Pages/Roles/RoleBulkAssignButton";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  try {
    const role = await Role.getById(id);
    if (!role) {
      notFound();
    }
    return {
      title: `${role.name} - ロール編集`,
      description: `${role.name}の詳細設定を行います`,
    };
  } catch {
    return {
      title: "ロール編集",
      description: "ロールの詳細設定を行います",
    };
  }
}

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let role: Role | null = null;
  try {
    role = await Role.getById(id);
  } catch (error) {
    if (error === ResourceApiErrors.ResourceNotFound) {
      notFound();
    }
  }
  if (!role) {
    notFound();
  }
  return (
    <Stack spacing={3}>
      <Breadcrumbs aria-label="breadcrumb">
        <Link underline="hover" color="inherit" href="/dashboard/roles">
          ロール管理
        </Link>
        <Typography sx={{ color: "text.primary" }}>編集</Typography>
      </Breadcrumbs>

      <Stack>
        <Typography variant="h4" gutterBottom>
          {role.name}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          ロールの詳細設定を行います。
        </Typography>
      </Stack>

      <Box>
        <RoleEditForm role={role.toJson()} />
      </Box>

      <Stack spacing={2}>
        <Typography variant="h5" gutterBottom>
          紐づいたユーザー
        </Typography>
        <Typography variant="body2" color="text.secondary" gutterBottom>
          このロールが割り当てられているユーザーの一覧です。
        </Typography>
        <RoleBulkAssignButton roleId={role.id} />
        <Suspense fallback={<div>Loading...</div>}>
          <RoleUsersDataGrid role={role} />
        </Suspense>
      </Stack>
    </Stack>
  );
}
