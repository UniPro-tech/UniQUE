import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const AuthorizationErrors = {
  AccessDenied: new UniQUE_Error(`アクセスが拒否されました。`, {
    code: "B0001",
    type: UniQUE_ErrorType.AUTHORIZATION_ERROR,
  }),
  Unauthorized: new UniQUE_Error(`認証されていません。`, {
    code: "B0002",
    type: UniQUE_ErrorType.AUTHORIZATION_ERROR,
  }),
};
