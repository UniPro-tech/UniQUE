import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const FrontendErrors = {
  UnhandledException: new UniQUE_Error(`予期せぬエラーが発生しました。`, {
    code: "E0001",
    type: UniQUE_ErrorType.FRONTEND_ERROR,
  }),
  NetworkError: new UniQUE_Error(`ネットワークエラーが発生しました。`, {
    code: "E0002",
    type: UniQUE_ErrorType.FRONTEND_ERROR,
  }),
  InvalidInput: new UniQUE_Error(`不正な入力です。`, {
    code: "E0003",
    type: UniQUE_ErrorType.FRONTEND_ERROR,
  }),
  TimeoutError: new UniQUE_Error(`処理がタイムアウトしました。`, {
    code: "E0004",
    type: UniQUE_ErrorType.FRONTEND_ERROR,
  }),
  SettingsLoadError: new UniQUE_Error(`設定の読み込みに失敗しました。`, {
    code: "E0005",
    type: UniQUE_ErrorType.FRONTEND_ERROR,
  }),
};
