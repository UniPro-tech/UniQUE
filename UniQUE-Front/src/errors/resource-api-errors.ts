import { UniQUE_Error, UniQUE_ErrorType } from "./base";

export const ResourceApiErrors = {
  ResourceNotFound: new UniQUE_Error(`リソースが見つかりません。`, {
    code: "R0001",
    type: UniQUE_ErrorType.RESOURCE_API_ERROR,
  }),
  ResourceAlreadyExists: new UniQUE_Error(`リソースは既に存在します。`, {
    code: "R0002",
    type: UniQUE_ErrorType.RESOURCE_API_ERROR,
  }),
  ResourceCreationFailed: new UniQUE_Error(`リソースの作成に失敗しました。`, {
    code: "R0003",
    type: UniQUE_ErrorType.RESOURCE_API_ERROR,
  }),
  ResourceUpdateFailed: new UniQUE_Error(`リソースの更新に失敗しました。`, {
    code: "R0004",
    type: UniQUE_ErrorType.RESOURCE_API_ERROR,
  }),
  ResourceDeletionFailed: new UniQUE_Error(`リソースの削除に失敗しました。`, {
    code: "R0005",
    type: UniQUE_ErrorType.RESOURCE_API_ERROR,
  }),
  CustomIdAlreadyExists: new UniQUE_Error(
    `ご指定のユーザーIDは既に使用されています。`,
    {
      code: "R1001",
      type: UniQUE_ErrorType.RESOURCE_API_ERROR,
    },
  ),
  ExternalEmailAlreadyExists: new UniQUE_Error(
    `このメールアドレスは既に登録されています。。`,
    {
      code: "R1002",
      type: UniQUE_ErrorType.RESOURCE_API_ERROR,
    },
  ),
  ApiServerInternalError: new UniQUE_Error(
    `APIサーバーでエラーが発生しました。`,
    {
      code: "R9006",
      type: UniQUE_ErrorType.RESOURCE_API_ERROR,
    },
  ),
};
