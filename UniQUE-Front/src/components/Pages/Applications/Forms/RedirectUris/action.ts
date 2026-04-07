"use server";

import { Application, type ApplicationData } from "@/classes/Application";

export const addRedirectUri = async (
  applicationData: ApplicationData,
  newUri: string,
) => {
  const application = Application.fromJson(applicationData);
  await application.addRedirectUri(newUri);
};

export const deleteRedirectUri = async (
  applicationData: ApplicationData,
  uriToDelete: string,
) => {
  const app = Application.fromJson(applicationData);
  await app.removeRedirectUri(uriToDelete);
};
