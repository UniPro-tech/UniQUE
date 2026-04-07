import type { SnackbarData } from "@/components/TemporarySnackProvider";
import {
  FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "./FrontendErrors";

export enum AuthenticationErrorCodes {
  MissingCredentials = "A0001",
  InvalidCredentials = "A0002",
  AccountLocked = "A0003",
  TokenExpired = "A0004",
  InsufficientPermissions = "A0005",
  InvalidEmailVerificationCode = "A0006",
  MigrationError = "A9001",
}

export const AuthenticationErrors = {
  MissingCredentials: new Error(
    `[${AuthenticationErrorCodes.MissingCredentials}] Missing credentials.`,
  ),
  InvalidCredentials: new Error(
    `[${AuthenticationErrorCodes.InvalidCredentials}] Invalid credentials.`,
  ),
  AccountLocked: new Error(
    `[${AuthenticationErrorCodes.AccountLocked}] Account is locked.`,
  ),
  TokenExpired: new Error(
    `[${AuthenticationErrorCodes.TokenExpired}] Token has expired.`,
  ),
  InsufficientPermissions: new Error(
    `[${AuthenticationErrorCodes.InsufficientPermissions}] Insufficient permissions.`,
  ),
  InvalidEmailVerificationCode: new Error(
    `[${AuthenticationErrorCodes.InvalidEmailVerificationCode}] Invalid email verification code.`,
  ),
  MigrationError: new Error(
    `[${AuthenticationErrorCodes.MigrationError}] Account migration error.`,
  ),
};

export function getAuthenticationErrorSnackbarData(
  error: AuthenticationErrorCodes,
): SnackbarData {
  switch (error) {
    case AuthenticationErrorCodes.MissingCredentials:
      return {
        message: `[${AuthenticationErrorCodes.MissingCredentials}] 認証情報が見つかりません。再度サインインしてください。`,
        variant: "error",
      };
    case AuthenticationErrorCodes.InvalidCredentials:
      return {
        message: `[${AuthenticationErrorCodes.InvalidCredentials}] 認証情報が無効です。再度サインインしてください。`,
        variant: "error",
      };
    case AuthenticationErrorCodes.AccountLocked:
      return {
        message: `[${AuthenticationErrorCodes.AccountLocked}] アカウントがロックされています。サポートにお問い合わせください。`,
        variant: "error",
      };
    case AuthenticationErrorCodes.TokenExpired:
      return {
        message: `[${AuthenticationErrorCodes.TokenExpired}] セッションの有効期限が切れました。再度サインインしてください。`,
        variant: "warning",
      };
    case AuthenticationErrorCodes.InsufficientPermissions:
      return {
        message: `[${AuthenticationErrorCodes.InsufficientPermissions}] 操作を実行する権限がありません。`,
        variant: "error",
      };
    case AuthenticationErrorCodes.InvalidEmailVerificationCode:
      return {
        message: `[${AuthenticationErrorCodes.InvalidEmailVerificationCode}] メール認証コードが無効です。再度お試しください。`,
        variant: "warning",
      };
    case AuthenticationErrorCodes.MigrationError:
      return {
        message: `[${AuthenticationErrorCodes.MigrationError}] アカウントの移行中にエラーが発生しました。ユーザー情報をお確かめの上、再度お試しください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}
