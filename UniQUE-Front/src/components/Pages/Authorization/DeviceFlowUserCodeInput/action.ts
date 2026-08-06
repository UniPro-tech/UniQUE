"use server";

import { AuthServerErrors } from "@/errors/AuthServerErrors";
import { createApiClient } from "@/libs/apiClient";

export const getAuthRequest = async (
  userCode: string,
): Promise<null | string> => {
  const apiClient = createApiClient(process.env.AUTH_API_URL);
  const res = await apiClient.get(`/internal/device-auth-requests/${userCode}`);
  if (!res.ok) {
    if (res.status === 404) {
      return null;
    }
    throw Error(AuthServerErrors.InternalServerError.message);
  }
  const data = await res.json();
  return data.id;
};
