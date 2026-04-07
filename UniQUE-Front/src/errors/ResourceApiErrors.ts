import type { SnackbarData } from "@/components/TemporarySnackProvider";
import {
  FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "./FrontendErrors";

export enum ResourceApiErrorCodes {
  ResourceNotFound = "R0001",
  ResourceAlreadyExists = "R0002",
  ResourceCreationFailed = "R0003",
  ResourceUpdateFailed = "R0004",
  ResourceDeletionFailed = "R0005",
  UsernameAlreadyExists = "R0006",
  EmailAlreadyExists = "R0007",
  ApiServerInternalError = "R1001",
}

export const ResourceApiErrors = {
  ResourceNotFound: new Error(
    `[${ResourceApiErrorCodes.ResourceNotFound}] Resource not found.`,
  ),
  ResourceAlreadyExists: new Error(
    `[${ResourceApiErrorCodes.ResourceAlreadyExists}] Resource already exists.`,
  ),
  ResourceCreationFailed: new Error(
    `[${ResourceApiErrorCodes.ResourceCreationFailed}] Resource creation failed.`,
  ),
  ResourceUpdateFailed: new Error(
    `[${ResourceApiErrorCodes.ResourceUpdateFailed}] Resource update failed.`,
  ),
  ResourceDeletionFailed: new Error(
    `[${ResourceApiErrorCodes.ResourceDeletionFailed}] Resource deletion failed.`,
  ),
  UsernameAlreadyExists: new Error(
    `[${ResourceApiErrorCodes.UsernameAlreadyExists}] Username already exists.`,
  ),
  EmailAlreadyExists: new Error(
    `[${ResourceApiErrorCodes.EmailAlreadyExists}] Email already exists.`,
  ),
  ApiServerInternalError: new Error(
    `[${ResourceApiErrorCodes.ApiServerInternalError}] API server internal error.`,
  ),
};

export function getResourceApiErrorSnackbarData(
  error: ResourceApiErrorCodes,
): SnackbarData {
  switch (error) {
    case ResourceApiErrorCodes.ResourceNotFound:
      return {
        message: `[${ResourceApiErrorCodes.ResourceNotFound}] リソースが見つかりません。再度お試しください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.ResourceAlreadyExists:
      return {
        message: `[${ResourceApiErrorCodes.ResourceAlreadyExists}] リソースは既に存在します。再度お試しください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.ResourceCreationFailed:
      return {
        message: `[${ResourceApiErrorCodes.ResourceCreationFailed}] リソースの作成に失敗しました。再度お試しください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.ResourceUpdateFailed:
      return {
        message: `[${ResourceApiErrorCodes.ResourceUpdateFailed}] リソースの更新に失敗しました。再度お試しください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.ResourceDeletionFailed:
      return {
        message: `[${ResourceApiErrorCodes.ResourceDeletionFailed}] リソースの削除に失敗しました。再度お試しください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.UsernameAlreadyExists:
      return {
        message: `[${ResourceApiErrorCodes.UsernameAlreadyExists}] このカスタムIDは既に使用されています。別のカスタムIDを入力してください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.EmailAlreadyExists:
      return {
        message: `[${ResourceApiErrorCodes.EmailAlreadyExists}] このメールアドレスは既に使用されています。別のメールアドレスを入力してください。`,
        variant: "error",
      };
    case ResourceApiErrorCodes.ApiServerInternalError:
      return {
        message: `[${ResourceApiErrorCodes.ApiServerInternalError}] APIサーバーでエラーが発生しました。時間をおいて再度お試しください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}
