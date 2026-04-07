import type { SnackbarData } from "@/components/TemporarySnackProvider";

export enum FrontendErrorCodes {
  UnhandledException = "E0001",
  NetworkError = "E0002",
  // ユーザーの入力に起因しないValidation Error
  InvalidInput = "E0003",
  TimeoutError = "E0004",
  SettingsLoadError = "E0005",
}

export const FrontendErrors = {
  UnhandledException: new Error(
    `[${FrontendErrorCodes.UnhandledException}] An unhandled exception occurred.`,
  ),
  NetworkError: new Error(
    `[${FrontendErrorCodes.NetworkError}] A network error occurred.`,
  ),
  InvalidInput: new Error(
    `[${FrontendErrorCodes.InvalidInput}] Invalid input provided.`,
  ),
  TimeoutError: new Error(
    `[${FrontendErrorCodes.TimeoutError}] The operation timed out.`,
  ),
  SettingsLoadError: new Error(
    `[${FrontendErrorCodes.SettingsLoadError}] Failed to load settings.`,
  ),
};

export function getFrontendErrorSnackbarData(
  error: FrontendErrorCodes,
): SnackbarData {
  switch (error) {
    case FrontendErrorCodes.UnhandledException:
      return {
        message: `[${FrontendErrorCodes.UnhandledException}] 不明なエラーが発生しました。再度お試しください。`,
        variant: "error",
      };
    case FrontendErrorCodes.NetworkError:
      return {
        message: `[${FrontendErrorCodes.NetworkError}] ネットワークエラーが発生しました。接続を確認してください。`,
        variant: "error",
      };
    case FrontendErrorCodes.InvalidInput:
      return {
        message: `[${FrontendErrorCodes.InvalidInput}] 無効な入力が提供されました。入力内容を確認してください。`,
        variant: "warning",
      };
    case FrontendErrorCodes.TimeoutError:
      return {
        message: `[${FrontendErrorCodes.TimeoutError}] 操作がタイムアウトしました。再度お試しください。`,
        variant: "warning",
      };
    case FrontendErrorCodes.SettingsLoadError:
      return {
        message: `[${FrontendErrorCodes.SettingsLoadError}] 設定の読み込みに失敗しました。ページを再読み込みしてください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}
