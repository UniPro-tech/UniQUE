"use server";
import "server-only";

import { createApiClient } from "@/libs/apiClient";

export const deniedAction = async (auth_request_id: string) => {
  const apiClientForAuthReq = createApiClient(process.env.AUTH_API_URL);
  await apiClientForAuthReq.delete(
    `/internal/auth-requests/${auth_request_id}`,
  );
};
