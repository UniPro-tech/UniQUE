import type { SnackbarData } from "@/components/TemporarySnackProvider";
import {
  FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "./FrontendErrors";

export enum AuthServerErrorCodes {
  InternalServerError = "D0001",
}

export const AuthServerErrors = {
  InternalServerError: new Error(
    `[${AuthServerErrorCodes.InternalServerError}] Internal server error on authentication server.`,
  ),
};

export function getAuthServerErrorSnackbarData(
  error: AuthServerErrorCodes,
): SnackbarData {
  switch (error) {
    case AuthServerErrorCodes.InternalServerError:
      return {
        message: `[${AuthServerErrorCodes.InternalServerError}] 認証サーバーで内部エラーが発生しました。しばらくしてから再度お試しください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}
