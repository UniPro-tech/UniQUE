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
  const session = await Session.getCurrent();

  if (!session) {
    const query = new URLSearchParams(params as Record<string, string>);
    const recirectpath = `/authorization?${query.toString()}`;
    redirect(
      `/signin?redirect=${encodeURIComponent(recirectpath)}`,
      RedirectType.replace,
    );
  }

  const sessionJson = session.toJson();
  if (!sessionJson) {
    const query = new URLSearchParams(params as Record<string, string>);
    const recirectpath = `/authorization?${query.toString()}`;
    redirect(
      `/signin?redirect=${encodeURIComponent(recirectpath)}`,
      RedirectType.replace,
    );
  }

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

  // 既存の同意があるかチェック – あればコンセント画面をスキップ
  // Resolve auth API URL from public or server env, fallback to localhost
  const resolvedAuthApiUrl =
    process.env.NEXT_PUBLIC_AUTH_API_URL || process.env.AUTH_API_URL;
  const authClient = createApiClient(resolvedAuthApiUrl);
  let consented = false;
  const consentedQuery = new URLSearchParams();
  try {
    const query = new URLSearchParams();
    query.append("user_id", sessionJson.userId);
    query.append("application_id", authReqData.client_id);
    query.append("scope", authReqData.scope);
    const consentsRes = await authClient.get(
      `/internal/consents?${query.toString()}`,
    );
    if (consentsRes.ok) {
      const consentsData = await consentsRes.json();
      const consents: { client_id?: string; application_id?: string }[] =
        Array.isArray(consentsData) ? consentsData : (consentsData.data ?? []);
      const hasConsent = consents.some(
        (c): boolean => c.application_id === authReqData.client_id,
      );

      if (hasConsent && authReqData.prompt === "none") {
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
      `${resolvedAuthApiUrl}/consented?${consentedQuery.toString()}`,
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
        action={`${resolvedAuthApiUrl}/authorization`}
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
