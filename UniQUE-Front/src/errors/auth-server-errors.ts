import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const AuthServerErrors = {
  InternalServerError: new UniQUE_Error(`内部エラーが発生しました。`, {
    code: "D0001",
    type: UniQUE_ErrorType.AUTH_SERVER_ERROR,
  }),
};
