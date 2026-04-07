/// ## Permissions
/// 権限ビットフラグの定義
/// 各ビットは特定の操作に対応し、組み合わせて使用することで複数の権限を表現できる.
/// 例えば、USER_READ | USER_CREATE はユーザーの読み取りと作成の両方の権限を持つことを意味する.
/// ## 設計方針
/// 0-7     = USER 系
/// 8-15    = CLIENT 系
/// 16-23   = SYSTEM 系
/// 24-31   = 予備
export const PermissionBitsFields = {
  // --- User Management ---
  // NOTE: もちろん自身のユーザー情報の更新・削除は別途許可される.
  // ここでは他者のユーザー情報に対する操作権限を定義する.
  USER_READ: 1n << 0n, // ユーザー読み取り
  USER_CREATE: 1n << 1n, // ユーザー作成
  USER_UPDATE: 1n << 2n, // ユーザー更新(パスワード無効化も含まれる)
  USER_DELETE: 1n << 3n, // ユーザー削除
  USER_DISABLE: 1n << 4n, // ユーザー無効化

  // --- Apps Management ---
  // NOTE: App は OAuth2 Client に相当する.
  // また、自身がOwnerである App に対しては別途許可される.
  // ここでは他者が所有する App に対する操作権限を定義する.
  APP_READ: 1n << 8n, // App読み取り
  // APP_CREATE = 1 << 9, // App作成 廃止
  APP_UPDATE: 1n << 10n, // App更新
  APP_DELETE: 1n << 11n, // App削除
  APP_SECRET_ROTATE: 1n << 12n, // Appシークレットの再発行

  // --- System / Config ---
  TOKEN_REVOKE: 1n << 16n, // トークンの取り消し
  AUDIT_READ: 1n << 18n, // 監査ログの読み取り
  CONFIG_UPDATE: 1n << 19n, //　 全体設定（認証フロー、署名鍵など）の変更
  KEY_MANAGE: 1n << 20n, // JWK鍵管理（追加／削除）

  // --- RBAC / Security ---
  ROLE_MANAGE: 1n << 24n, // RBACロール自体の作成／編集／削除
  PERMISSION_MANAGE: 1n << 25n, // RBAC権限の割り当て管理
  SESSION_MANAGE: 1n << 26n, // セッション管理（強制ログアウトなど）
  MFA_MANAGE: 1n << 27n, // 多要素認証の管理(リセットなど)

  // --- Announcements (32-39) ---
  ANNOUNCEMENT_CREATE: 1n << 32n, // お知らせの作成 (1 << 32)
  ANNOUNCEMENT_UPDATE: 1n << 33n, // お知らせの編集 (1 << 33)
  ANNOUNCEMENT_DELETE: 1n << 34n, // お知らせの削除 (1 << 34)
  ANNOUNCEMENT_PIN: 1n << 35n, // お知らせのピン留め (1 << 35)
};

export enum PermissionTexts {
  USER_READ = "ユーザー読み取り",
  USER_CREATE = "ユーザー作成",
  USER_UPDATE = "ユーザー更新",
  USER_DELETE = "ユーザー削除",
  USER_DISABLE = "ユーザー無効化",
  APP_READ = "アプリ読み取り",
  // "APP_CREATE" = "アプリ作成", --- IGNORE ---
  APP_UPDATE = "アプリ更新",
  APP_DELETE = "アプリ削除",
  APP_SECRET_ROTATE = "アプリシークレットの再発行",
  TOKEN_REVOKE = "トークン取り消し",
  AUDIT_READ = "監査ログ読み取り",
  CONFIG_UPDATE = "全体設定変更",
  KEY_MANAGE = "JWK鍵管理",
  ROLE_MANAGE = "ロール管理",
  PERMISSION_MANAGE = "権限管理",
  SESSION_MANAGE = "セッション管理",
  MFA_MANAGE = "多要素認証管理",
  ANNOUNCEMENT_CREATE = "お知らせ作成",
  ANNOUNCEMENT_UPDATE = "お知らせ編集",
  ANNOUNCEMENT_DELETE = "お知らせ削除",
  ANNOUNCEMENT_PIN = "お知らせピン留め",
}

export const ConvertPermissionBitsToText = (
  permissionBits: bigint,
): string[] => {
  const permissionTexts: string[] = [];
  for (const [key, value] of Object.entries(PermissionBitsFields)) {
    if ((permissionBits & value) === value) {
      permissionTexts.push(
        PermissionTexts[key as keyof typeof PermissionTexts],
      );
    }
  }
  return permissionTexts;
};
