import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import type { VariantType } from "notistack";
import { AuthorizationPageMode } from "@/components/Pages/Authentication";
import AuthenticationPage, {
  type AuthorizationFormState,
} from "@/components/Pages/Authentication/Client";
import TemporarySnackProvider, {
  type SnackbarData,
} from "@/components/TemporarySnackProvider";
import {
  type AuthenticationErrorCodes,
  getAuthenticationErrorSnackbarData,
} from "@/errors/AuthenticationErrors";
import {
  type AuthServerErrorCodes,
  getAuthServerErrorSnackbarData,
} from "@/errors/AuthServerErrors";
import {
  type FormRequestErrorCodes,
  getFormRequestErrorSnackbarData,
} from "@/errors/FormRequestErrors";
import {
  type FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "@/errors/FrontendErrors";

export const metadata = {
  title: "サインイン",
  description: "UniQUEのサインインページです。UniQUEにサインインします。",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{
    name?: string;
    username?: string;
    email?: string;
    externalEmail?: string;
    agreeToTerms?: string;
    rememberMe?: string;
    migration?: string;
    signouted?: string;
    redirect?: string;
    error?:
      | AuthenticationErrorCodes
      | FormRequestErrorCodes
      | AuthServerErrorCodes
      | FrontendErrorCodes;
  }>;
}) {
  const {
    migration,
    error,
    signouted,
    name,
    username,
    email,
    externalEmail,
    agreeToTerms,
    rememberMe,
    redirect: redirectPath,
  } = await searchParams;
  const initState: AuthorizationFormState = {
    name,
    username,
    email,
    externalEmail,
    agreeToTerms: agreeToTerms === "1",
    rememberMe: rememberMe === "1",
  };
  const snacks: SnackbarData[] = [
    ...(signouted
      ? [
          {
            message: `サインアウトしました。`,
            variant: "success" as VariantType,
          },
        ]
      : []),
    ...(migration
      ? [
          {
            message: `アカウントの移行が完了しました。サインインしてください。`,
            variant: "success" as VariantType,
          },
        ]
      : []),
    ...(error?.startsWith("A")
      ? [getAuthenticationErrorSnackbarData(error as AuthenticationErrorCodes)]
      : error?.startsWith("F")
        ? [getFormRequestErrorSnackbarData(error as FormRequestErrorCodes)]
        : error?.startsWith("D")
          ? [getAuthServerErrorSnackbarData(error as AuthServerErrorCodes)]
          : error?.startsWith("E")
            ? [getFrontendErrorSnackbarData(error as FrontendErrorCodes)]
            : []),
  ];

  const cookieStore = await cookies();
  const pending_mfa_remember =
    cookieStore.get("pending_mfa_remember")?.value === "1";
  const pending_mfa_user = cookieStore.get("pending_mfa_user")?.value;
  if (!pending_mfa_user) {
    redirect("/signin");
  }
  initState.username = pending_mfa_user;
  initState.rememberMe = pending_mfa_remember;
  return (
    <>
      <TemporarySnackProvider snacks={snacks} />
      <AuthenticationPage
        initFormState={initState}
        mode={AuthorizationPageMode.MFA}
        redirectTo={redirectPath}
      />
    </>
  );
}
