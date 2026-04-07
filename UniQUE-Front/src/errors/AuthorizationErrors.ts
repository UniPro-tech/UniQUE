import type { SnackbarData } from "@/components/TemporarySnackProvider";
import {
  FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "./FrontendErrors";

export enum AuthorizationErrorCodes {
  AccessDenied = "B0001",
  Unauthorized = "B0002",
}

export const AuthorizationErrors = {
  AccessDenied: new Error(
    `[${AuthorizationErrorCodes.AccessDenied}] Access denied.`,
  ),
  Unauthorized: new Error(
    `[${AuthorizationErrorCodes.Unauthorized}] Unauthorized access.`,
  ),
};

export function getAuthorizationErrorSnackbarData(
  error: AuthorizationErrorCodes,
): SnackbarData {
  switch (error) {
    case AuthorizationErrorCodes.AccessDenied:
      return {
        message: `[${AuthorizationErrorCodes.AccessDenied}] アクセスが拒否されました。権限を確認してください。`,
        variant: "error",
      };
    case AuthorizationErrorCodes.Unauthorized:
      return {
        message: `[${AuthorizationErrorCodes.Unauthorized}] 認証情報が無効です。再度サインインしてください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}
