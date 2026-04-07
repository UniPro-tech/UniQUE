import { PermissionBitsFields } from "@/constants/Permission";

// 権限のグループ化
export const PermissionGroups = [
  {
    title: "ユーザー管理",
    permissions: [
      PermissionBitsFields.USER_READ,
      PermissionBitsFields.USER_CREATE,
      PermissionBitsFields.USER_UPDATE,
      PermissionBitsFields.USER_DELETE,
      PermissionBitsFields.USER_DISABLE,
    ],
  },
  {
    title: "アプリケーション管理",
    permissions: [
      PermissionBitsFields.APP_READ,
      PermissionBitsFields.APP_UPDATE,
      PermissionBitsFields.APP_DELETE,
      PermissionBitsFields.APP_SECRET_ROTATE,
    ],
  },
  {
    title: "お知らせ管理",
    permissions: [
      PermissionBitsFields.ANNOUNCEMENT_CREATE,
      PermissionBitsFields.ANNOUNCEMENT_UPDATE,
      PermissionBitsFields.ANNOUNCEMENT_DELETE,
      PermissionBitsFields.ANNOUNCEMENT_PIN,
    ],
  },
  {
    title: "システム管理",
    permissions: [
      PermissionBitsFields.TOKEN_REVOKE,
      PermissionBitsFields.AUDIT_READ,
      PermissionBitsFields.CONFIG_UPDATE,
      PermissionBitsFields.KEY_MANAGE,
    ],
  },
  {
    title: "RBAC・セキュリティ",
    permissions: [
      PermissionBitsFields.ROLE_MANAGE,
      PermissionBitsFields.PERMISSION_MANAGE,
      PermissionBitsFields.SESSION_MANAGE,
      PermissionBitsFields.MFA_MANAGE,
    ],
  },
];
