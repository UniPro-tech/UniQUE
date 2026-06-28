import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const FormRequestErrors = {
  InvalidFormData: new UniQUE_Error(`不正なデータです。`, {
    code: "F0001",
    type: UniQUE_ErrorType.FORM_REQUEST_ERROR,
  }),
  MissingRequiredFields: new UniQUE_Error(`必須データが欠損しています。`, {
    code: "F0002",
    type: UniQUE_ErrorType.FORM_REQUEST_ERROR,
  }),
  ExceededFieldLength: new UniQUE_Error(`項目の最大量に達しました。`, {
    code: "F0003",
    type: UniQUE_ErrorType.FORM_REQUEST_ERROR,
  }),
  InvalidFieldFormat: new UniQUE_Error(`フィールドの形式が不正です。`, {
    code: "F0004",
    type: UniQUE_ErrorType.FORM_REQUEST_ERROR,
  }),
};
