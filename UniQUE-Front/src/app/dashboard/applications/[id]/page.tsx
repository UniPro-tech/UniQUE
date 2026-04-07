import { Breadcrumbs, Link, Stack, Typography } from "@mui/material";
import { notFound } from "next/navigation";
import { Application } from "@/classes/Application";
import ApplicationEditForm from "@/components/Pages/Applications/Forms/ApplicationEditForm";
import RedirectUris from "@/components/Pages/Applications/Forms/RedirectUris";
import { ResourceApiErrors } from "@/errors/ResourceApiErrors";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  try {
    const app = await Application.getById(id);
    return {
      title: `${app.name} - アプリケーション編集`,
      description: `${app.name}の詳細設定を行います`,
    };
  } catch {
    return {
      title: "アプリケーション編集",
      description: "アプリケーションの詳細設定を行います",
    };
  }
}

export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  let application: Application | null = null;

  try {
    application = await Application.getById(id);
  } catch (error) {
    if (error === ResourceApiErrors.ResourceNotFound) {
      notFound();
    }
  }

  if (!application) {
    notFound();
  }

  return (
    <Stack spacing={3}>
      <Breadcrumbs aria-label="breadcrumb">
        <Link underline="hover" color="inherit" href="/dashboard/applications">
          アプリケーション管理
        </Link>
        <Typography sx={{ color: "text.primary" }}>編集</Typography>
      </Breadcrumbs>

      <Stack>
        <Typography variant="h4" gutterBottom>
          {application.name}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          アプリケーションの詳細設定を行います。
        </Typography>
      </Stack>

      <ApplicationEditForm application={await application.toJson()} />
      <RedirectUris application={application} />
    </Stack>
  );
}
