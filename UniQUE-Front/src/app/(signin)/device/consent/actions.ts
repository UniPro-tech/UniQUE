"use server";
import "server-only";

import { createApiClient } from "@/libs/apiClient";

export const deniedAction = async (auth_request_id: string) => {
  const apiClientForAuthReq = createApiClient(process.env.AUTH_API_URL);
  // auth_request_idがULIDか検証する
  if (!/^[0-9A-HJKMNP-TV-Z]{26}$/.test(auth_request_id)) {
    throw new Error("Invalid auth_request_id");
  }

  const response = await apiClientForAuthReq.delete(
    `/internal/auth-requests/${auth_request_id}`,
  );

  if (!response.ok) {
    throw new Error("Failed to deny authorization request");
  }
};
