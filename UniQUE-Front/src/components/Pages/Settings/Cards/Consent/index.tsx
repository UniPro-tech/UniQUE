import type { UserData } from "@/classes/types/User";
import { createApiClient } from "@/libs/apiClient";
import { toCamelcase } from "@/libs/snakeCamelUtil";
import { revokeConsentAction } from "./actions";
import ConsentSettingsCardClient, { type ConsentDTO } from "./Client";

export default async function ConsentSettingsCard({
  user,
}: {
  user: UserData;
}) {
  const authApiUrl = process.env.AUTH_API_URL;
  const authClient = createApiClient(authApiUrl);
  let consents: ConsentDTO[] = [];

  try {
    const requestQuery = new URLSearchParams({
      user_id: user.id,
    });
    const res = await authClient.get(
      `/internal/consents?${requestQuery.toString()}`,
    );
    if (res.ok) {
      const data = await res.json();
      const raw = Array.isArray(data) ? data : (data.data ?? []);
      consents = raw.map(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (c: any) => toCamelcase(c) as ConsentDTO,
      );
    }
  } catch {
    // フェッチ失敗時は空リスト
  }

  // アプリ名解決: Resource API から application 情報を取得
  const resourceApiUrl =
    process.env.RESOURCE_API_URL || process.env.NEXT_PUBLIC_RESOURCE_API_URL;
  const resourceClient = createApiClient(resourceApiUrl);
  const resolvedConsents = await Promise.all(
    consents.map(async (c) => {
      const appId = c.clientId || c.applicationId;
      if (!appId) return c;
      try {
        const appRes = await resourceClient.get(
          `/applications/${encodeURIComponent(appId)}`,
        );
        if (appRes.ok) {
          const app = await appRes.json();
          return {
            ...c,
            applicationName: app.name ?? appId,
            applicationDescription: app.description,
            applicationWebsiteUrl: app.website_url,
            applicationPrivacyPolicyUrl: app.privacy_policy_url,
          };
        }
      } catch {
        // ignore
      }
      return { ...c, applicationName: appId };
    }),
  );

  return (
    <ConsentSettingsCardClient
      consents={resolvedConsents}
      revokeConsent={revokeConsentAction}
    />
  );
}
