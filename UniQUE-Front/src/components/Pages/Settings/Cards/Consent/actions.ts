"use server";

import { createApiClient } from "@/libs/apiClient";

export async function revokeConsentAction(consentId: string): Promise<boolean> {
  const authApiUrl = process.env.AUTH_API_URL;
  try {
    const client = createApiClient(authApiUrl);
    const res = await client.delete(
      `/internal/consents/${encodeURIComponent(consentId)}`,
    );
    return res.ok;
  } catch {
    return false;
  }
}
