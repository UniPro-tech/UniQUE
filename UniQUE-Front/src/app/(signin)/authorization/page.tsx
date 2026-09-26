import { cookies } from "next/headers";
import { RedirectType, redirect } from "next/navigation";
import { Application } from "@/classes/Application";
import { Session } from "@/classes/Session";
import ConsentCard from "@/components/Pages/Authorization/ConsentCard";
import TemporarySnackProvider, {
  type SnackbarData,
} from "@/components/TemporarySnackProvider";
import { createApiClient } from "@/libs/apiClient";

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{
    auth_request_id?: string;
    error?: string;
  }>;
}) {
  const params = await searchParams;
  const { auth_request_id, error } = params;

  if (error) {
    return (
      <main style={{ padding: 24 }}>
        <h1>Authorization Error</h1>
        <p>
          {error === "forbidden_scope"
            ? "このスコープに対する認可を行う権限がありません。"
            : error}
        </p>
      </main>
    );
  }

  if (!auth_request_id) {
    return (
      <main style={{ padding: 24 }}>
        <h1>Bad Request</h1>
        <p>不正なリクエストです。</p>
      </main>
    );
  }

  // auth_request_idから認可リクエスト情報を取得
  const apiClientForAuthReq = createApiClient(process.env.AUTH_API_URL);
  const authReqRes = await apiClientForAuthReq.get(
    `/internal/auth-requests/${encodeURIComponent(auth_request_id || "")}`,
  );
  const snacks: SnackbarData[] = [];
  if (!authReqRes.ok) {
    snacks.push({
      message: "不正なリクエストです。AuthRequestの取得に失敗しました。",
      variant: "error",
    });
    return (
      <>
        <TemporarySnackProvider snacks={snacks} />
        <main style={{ padding: 24 }}>
          <h1>Bad Request</h1>
          <p>不正なリクエストです。</p>
        </main>
      </>
    );
  }
  const authReqData = (await authReqRes.json()) as AuthorizationResponse;
  const session = await Session.getCurrent();

  if (!session?.toJson()) {
    if (authReqData.prompt === "none") {
      const redirectUrl = new URL(authReqData.redirect_uri);
      redirectUrl.searchParams.set("error", "login_required");
      if (authReqData.state) {
        redirectUrl.searchParams.set("state", authReqData.state);
      }
      redirect(redirectUrl.toString(), RedirectType.replace);
    }

    const query = new URLSearchParams(params as Record<string, string>);
    const recirectpath = `/authorization?${query.toString()}`;
    redirect(
      `/signin?redirect=${encodeURIComponent(recirectpath)}`,
      RedirectType.replace,
    );
  }

  const sessionJson = session.toJson();

  const app = await Application.getById(authReqData.client_id);
  if (!app) {
    snacks.push({ message: "不正なクライアントIDです。", variant: "error" });
    return (
      <>
        <TemporarySnackProvider snacks={snacks} />
        <main style={{ padding: 24 }}>
          <h1>Bad Request</h1>
          <p>不正なクライアントIDです。</p>
        </main>
      </>
    );
  }

  const cookieStore = await cookies();
  const COOKIE_NAME = "session_jwt";
  const jwtToken = cookieStore.get(COOKIE_NAME)?.value;
  if (!jwtToken) {
    const query = new URLSearchParams(params as Record<string, string>);
    const recirectpath = `/authorization?${query.toString()}`;
    redirect(
      `/signin?redirect=${encodeURIComponent(recirectpath)}`,
      RedirectType.replace,
    );
  }

  // Server-side API calls must use the Docker-internal URL. Browser redirects
  // and form actions use the public URL instead.
  const authApiUrl =
    process.env.AUTH_API_URL || process.env.NEXT_PUBLIC_AUTH_API_URL;
  const publicAuthApiUrl = process.env.NEXT_PUBLIC_AUTH_API_URL || authApiUrl;
  const authClient = createApiClient(authApiUrl);
  let consented = false;
  const consentedQuery = new URLSearchParams();
  try {
    const query = new URLSearchParams();
    query.append("user_id", sessionJson.userId);
    query.append("application_id", authReqData.client_id);
    const consentsRes = await authClient.get(
      `/internal/consents?${query.toString()}`,
    );
    if (consentsRes.ok) {
      const consentsData = await consentsRes.json();
      const consents: { application_id?: string; scope?: string }[] =
        Array.isArray(consentsData) ? consentsData : (consentsData.data ?? []);
      const requestedScopes = new Set(
        authReqData.scope.split(/\s+/).filter(Boolean),
      );
      const hasConsent = consents.some((c): boolean => {
        if (c.application_id !== authReqData.client_id) return false;
        const consentedScopes = new Set(c.scope?.split(/\s+/).filter(Boolean));
        return [...requestedScopes].every((scope) =>
          consentedScopes.has(scope),
        );
      });

      if (!hasConsent && authReqData.prompt === "none") {
        consentRequired = true;
      }

      // Reuse a stored consent unless the client explicitly requests a new
      // approval screen with prompt=consent.
      if (hasConsent && authReqData.prompt !== "consent") {
        // 同意済みであればconsentedにする
        const query = new URLSearchParams();
        query.append("user_id", sessionJson.userId);
        query.append("application_id", authReqData.client_id);
        query.append("scope", authReqData.scope);
        const consentRes = await authClient.post(
          `/internal/auth-requests/${auth_request_id}/consented?${query.toString()}`,
        );
        if (!consentRes.ok) {
          throw new Error("Failed to create consent");
        }
        // リダイレクト
        consentedQuery.append("authorization_id", auth_request_id);
        consented = true;
      }
    }
  } catch {
    // 同意チェックに失敗した場合はフォールスルーして同意画面を表示
  }

  if (consented) {
    redirect(
      `${publicAuthApiUrl}/consented?${consentedQuery.toString()}`,
      RedirectType.push,
    );
  }

  return (
    <>
      <TemporarySnackProvider snacks={snacks} />
      <ConsentCard
        app={await app.toJson()}
        user={(await session.getUser()).toJson()}
        scope={authReqData.scope}
        jwt={jwtToken}
        redirect_uri={authReqData.redirect_uri}
        state={authReqData.state}
        auth_request_id={auth_request_id}
        action={`${publicAuthApiUrl}/authorization`}
      />
    </>
  );
}

interface AuthorizationResponse {
  client_id: string;
  redirect_uri: string;
  scope: string;
  state?: string;
  response_type?: string;
  prompt?: string;
  nonce?: string;
  code_challenge?: string;
  code_challenge_method?: string;
}
