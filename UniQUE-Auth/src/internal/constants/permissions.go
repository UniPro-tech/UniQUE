package constants

// Permission defines the permission bits for RBAC
// フロントエンドのPermissionBitsFieldsと一致させる
type Permission int64

const (
	// --- User Management (0-7) ---
	// NOTE: 自身のユーザー情報の更新・削除は別途許可される
	// ここでは他者のユーザー情報に対する操作権限を定義する
	USER_READ    Permission = 1 << 0 // ユーザー読み取り
	USER_CREATE  Permission = 1 << 1 // ユーザー作成
	USER_UPDATE  Permission = 1 << 2 // ユーザー更新(パスワード無効化も含まれる)
	USER_DELETE  Permission = 1 << 3 // ユーザー削除
	USER_DISABLE Permission = 1 << 4 // ユーザー無効化

	// --- Apps Management (8-15) ---
	// NOTE: App は OAuth2 Client に相当する
	// また、自身がOwnerである App に対しては別途許可される
	// ここでは他者が所有する App に対する操作権限を定義する
	APP_READ          Permission = 1 << 8  // App読み取り
	APP_UPDATE        Permission = 1 << 10 // App更新
	APP_DELETE        Permission = 1 << 11 // App削除
	APP_SECRET_ROTATE Permission = 1 << 12 // Appシークレットの再発行

	// --- System / Config (16-23) ---
	TOKEN_REVOKE  Permission = 1 << 16 // トークンの取り消し
	AUDIT_READ    Permission = 1 << 18 // 監査ログの読み取り
	CONFIG_UPDATE Permission = 1 << 19 // 全体設定（認証フロー、署名鍵など）の変更
	KEY_MANAGE    Permission = 1 << 20 // JWK鍵管理（追加／削除）

	// --- RBAC / Security (24-31) ---
	ROLE_MANAGE       Permission = 1 << 24 // RBACロール自体の作成／編集／削除
	PERMISSION_MANAGE Permission = 1 << 25 // RBAC権限の割り当て管理
	SESSION_MANAGE    Permission = 1 << 26 // セッション管理（強制ログアウトなど）
	MFA_MANAGE        Permission = 1 << 27 // 多要素認証の管理(リセットなど)

	// --- Announcements (32-39) ---
	ANNOUNCEMENT_CREATE Permission = 1 << 32 // お知らせの作成
	ANNOUNCEMENT_UPDATE Permission = 1 << 33 // お知らせの編集
	ANNOUNCEMENT_DELETE Permission = 1 << 34 // お知らせの削除
	ANNOUNCEMENT_PIN    Permission = 1 << 35 // お知らせのピン留め

	// --- Special Permissions ---
	// システム管理者（すべての権限）
	ADMIN Permission = -1
)

// ScopeRequirementsPermissionMap maps OAuth2/OIDC スコープ名を、
// 内部の権限ビットマスクへ紐付ける。権限が不要なスコープは 0 を設定する。
var ScopeRequirementsPermissionMap = map[string]Permission{
	"openid":               Permission(0),
	"profile":              Permission(0),
	"email":                Permission(0),
	"announcements.write":  ANNOUNCEMENT_CREATE | ANNOUNCEMENT_UPDATE,
	"announcements.delete": ANNOUNCEMENT_DELETE,
}

// HasPermission checks if the permission bitmask contains the required permission
func (p Permission) HasPermission(required Permission) bool {
	// ADMIN権限を持っている場合は全ての権限を持つ
	if p == ADMIN {
		return true
	}
	return (p & required) == required
}

// PermissionNames maps permission bits to human-readable names
var PermissionNames = map[Permission]string{
	USER_READ:           "ユーザー読み取り",
	USER_CREATE:         "ユーザー作成",
	USER_UPDATE:         "ユーザー更新",
	USER_DELETE:         "ユーザー削除",
	USER_DISABLE:        "ユーザー無効化",
	APP_READ:            "アプリ読み取り",
	APP_UPDATE:          "アプリ更新",
	APP_DELETE:          "アプリ削除",
	APP_SECRET_ROTATE:   "アプリシークレット再発行",
	TOKEN_REVOKE:        "トークン取り消し",
	AUDIT_READ:          "監査ログ読み取り",
	CONFIG_UPDATE:       "全体設定変更",
	KEY_MANAGE:          "JWK鍵管理",
	ROLE_MANAGE:         "ロール管理",
	PERMISSION_MANAGE:   "権限管理",
	SESSION_MANAGE:      "セッション管理",
	MFA_MANAGE:          "多要素認証管理",
	ANNOUNCEMENT_CREATE: "お知らせ作成",
	ANNOUNCEMENT_UPDATE: "お知らせ編集",
	ANNOUNCEMENT_DELETE: "お知らせ削除",
	ANNOUNCEMENT_PIN:    "お知らせピン留め",
	ADMIN:               "システム管理者",
}

// GetPermissionName returns the human-readable name for a permission
func GetPermissionName(p Permission) string {
	if name, ok := PermissionNames[p]; ok {
		return name
	}
	return "不明な権限"
}

// GetPermissionsText converts a permission bitmask to a list of text descriptions
func GetPermissionsText(bitmask int64) []string {
	var permissions []string
	p := Permission(bitmask)

	if p == ADMIN {
		return []string{"システム管理者（全権限）"}
	}

	for perm, name := range PermissionNames {
		if perm == ADMIN {
			continue
		}
		if p.HasPermission(perm) {
			permissions = append(permissions, name)
		}
	}

	return permissions
}
