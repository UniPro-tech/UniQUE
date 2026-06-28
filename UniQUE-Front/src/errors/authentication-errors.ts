import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const AuthenticationErrors = {
  MissingCredentials: new UniQUE_Error(`認証情報がありません。`, {
    code: "F0001",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  InvalidCredentials: new UniQUE_Error(`認証情報が間違っています。`, {
    code: "F0002",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  AccountLocked: new UniQUE_Error(`アカウントがロックされています。`, {
    code: "F0003",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  TokenExpired: new UniQUE_Error(`認証情報の有効期限が切れています。`, {
    code: "F0004",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  InsufficientPermissions: new UniQUE_Error(`権限が不足しています。`, {
    code: "F0005",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  InvalidVerificationCode: new UniQUE_Error(`認証コードが誤っています。`, {
    code: "F0006",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
  MigrationError: new UniQUE_Error(`アカウント移行に失敗しました。`, {
    code: "F1001",
    type: UniQUE_ErrorType.AUTHENTICATION_ERROR,
  }),
};
