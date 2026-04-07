import type { SnackbarData } from "@/components/TemporarySnackProvider";
import {
  FrontendErrorCodes,
  getFrontendErrorSnackbarData,
} from "./FrontendErrors";

export enum FormRequestErrorCodes {
  InvalidFormData = "F0001",
  MissingRequiredFields = "F0002",
  ExceededFieldLength = "F0003",
  InvalidFieldFormat = "F0004",
  CSRFTokenMismatch = "F0005",
}

export function getFormRequestErrorSnackbarData(
  error: FormRequestErrorCodes,
): SnackbarData {
  switch (error) {
    case FormRequestErrorCodes.InvalidFormData:
      return {
        message: `[${FormRequestErrorCodes.InvalidFormData}] フォームのデータが不正です。再度お試しください。`,
        variant: "error",
      };
    case FormRequestErrorCodes.MissingRequiredFields:
      return {
        message: `[${FormRequestErrorCodes.MissingRequiredFields}] 必須項目が未入力です。入力内容を確認してください。`,
        variant: "error",
      };
    case FormRequestErrorCodes.ExceededFieldLength:
      return {
        message: `[${FormRequestErrorCodes.ExceededFieldLength}] 入力内容が長すぎます。文字数を確認してください。`,
        variant: "error",
      };
    case FormRequestErrorCodes.InvalidFieldFormat:
      return {
        message: `[${FormRequestErrorCodes.InvalidFieldFormat}] 入力内容の形式が不正です。入力内容を確認してください。`,
        variant: "error",
      };
    case FormRequestErrorCodes.CSRFTokenMismatch:
      return {
        message: `[${FormRequestErrorCodes.CSRFTokenMismatch}] 不正なリクエストです。ページをリロードしてください。`,
        variant: "error",
      };
    default:
      return getFrontendErrorSnackbarData(
        FrontendErrorCodes.UnhandledException,
      );
  }
}

export const FormRequestErrors = {
  InvalidFormData: new Error(
    `[${FormRequestErrorCodes.InvalidFormData}] Invalid form data.`,
  ),
  MissingRequiredFields: new Error(
    `[${FormRequestErrorCodes.MissingRequiredFields}] Missing required fields.`,
  ),
  ExceededFieldLength: new Error(
    `[${FormRequestErrorCodes.ExceededFieldLength}] Exceeded field length.`,
  ),
  InvalidFieldFormat: new Error(
    `[${FormRequestErrorCodes.InvalidFieldFormat}] Invalid field format.`,
  ),
  CSRFTokenMismatch: new Error(
    `[${FormRequestErrorCodes.CSRFTokenMismatch}] CSRF token mismatch.`,
  ),
};
