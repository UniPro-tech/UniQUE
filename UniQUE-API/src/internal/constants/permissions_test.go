package constants_test

import (
	"slices"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/constants"
)

func TestHasPermission_SinglePermission(t *testing.T) {
	tests := []struct {
		name     string
		p        constants.Permission
		required constants.Permission
		want     bool
	}{
		{
			name:     "has USER_READ",
			p:        constants.USER_READ,
			required: constants.USER_READ,
			want:     true,
		},
		{
			name:     "does not have USER_CREATE",
			p:        constants.USER_READ,
			required: constants.USER_CREATE,
			want:     false,
		},
		{
			name:     "has USER_UPDATE",
			p:        constants.USER_UPDATE,
			required: constants.USER_UPDATE,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.HasPermission(tt.required); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermission_MultiplePermissions(t *testing.T) {
	tests := []struct {
		name     string
		p        constants.Permission
		required constants.Permission
		want     bool
	}{
		{
			name:     "has both USER_READ and USER_CREATE",
			p:        constants.USER_READ | constants.USER_CREATE,
			required: constants.USER_READ,
			want:     true,
		},
		{
			name:     "has both but check for third permission",
			p:        constants.USER_READ | constants.USER_CREATE,
			required: constants.USER_UPDATE,
			want:     false,
		},
		{
			name:     "has all three user management permissions",
			p:        constants.USER_READ | constants.USER_CREATE | constants.USER_UPDATE,
			required: constants.USER_CREATE,
			want:     true,
		},
		{
			name:     "cross-category permissions",
			p:        constants.USER_READ | constants.APP_UPDATE | constants.TOKEN_REVOKE,
			required: constants.APP_UPDATE,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.HasPermission(tt.required); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermission_AdminPermission(t *testing.T) {
	tests := []struct {
		name     string
		required constants.Permission
	}{
		{name: "ADMIN has USER_READ", required: constants.USER_READ},
		{name: "ADMIN has USER_DELETE", required: constants.USER_DELETE},
		{name: "ADMIN has APP_SECRET_ROTATE", required: constants.APP_SECRET_ROTATE},
		{name: "ADMIN has TOKEN_REVOKE", required: constants.TOKEN_REVOKE},
		{name: "ADMIN has ROLE_MANAGE", required: constants.ROLE_MANAGE},
		{name: "ADMIN has ANNOUNCEMENT_PIN", required: constants.ANNOUNCEMENT_PIN},
		{name: "ADMIN has any random permission", required: constants.Permission(1 << 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constants.ADMIN.HasPermission(tt.required); !got {
				t.Errorf("ADMIN.HasPermission(%v) = %v, want true", tt.required, got)
			}
		})
	}
}

func TestHasPermission_ZeroPermission(t *testing.T) {
	zeroPerm := constants.Permission(0)

	tests := []struct {
		name     string
		required constants.Permission
	}{
		{name: "zero permission does not have USER_READ", required: constants.USER_READ},
		{name: "zero permission does not have ADMIN", required: constants.ADMIN},
		{name: "zero permission does not have APP_READ", required: constants.APP_READ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zeroPerm.HasPermission(tt.required); got {
				t.Errorf("Permission(0).HasPermission(%v) = %v, want false", tt.required, got)
			}
		})
	}
}

func TestGetPermissionName_DefinedPermissions(t *testing.T) {
	tests := []struct {
		perm constants.Permission
		want string
	}{
		{constants.USER_READ, "ユーザー読み取り"},
		{constants.USER_CREATE, "ユーザー作成"},
		{constants.USER_UPDATE, "ユーザー更新"},
		{constants.USER_DELETE, "ユーザー削除"},
		{constants.USER_DISABLE, "ユーザー無効化"},
		{constants.APP_READ, "アプリ読み取り"},
		{constants.APP_UPDATE, "アプリ更新"},
		{constants.APP_DELETE, "アプリ削除"},
		{constants.APP_SECRET_ROTATE, "アプリシークレット再発行"},
		{constants.TOKEN_REVOKE, "トークン取り消し"},
		{constants.AUDIT_READ, "監査ログ読み取り"},
		{constants.CONFIG_UPDATE, "全体設定変更"},
		{constants.KEY_MANAGE, "JWK鍵管理"},
		{constants.ROLE_MANAGE, "ロール管理"},
		{constants.PERMISSION_MANAGE, "権限管理"},
		{constants.SESSION_MANAGE, "セッション管理"},
		{constants.MFA_MANAGE, "多要素認証管理"},
		{constants.ANNOUNCEMENT_CREATE, "お知らせ作成"},
		{constants.ANNOUNCEMENT_UPDATE, "お知らせ編集"},
		{constants.ANNOUNCEMENT_DELETE, "お知らせ削除"},
		{constants.ANNOUNCEMENT_PIN, "お知らせピン留め"},
		{constants.ADMIN, "システム管理者"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := constants.GetPermissionName(tt.perm); got != tt.want {
				t.Errorf("GetPermissionName(%v) = %q, want %q", tt.perm, got, tt.want)
			}
		})
	}
}

func TestGetPermissionName_UndefinedPermission(t *testing.T) {
	undefinedPerm := constants.Permission(1 << 50)
	want := "不明な権限"

	if got := constants.GetPermissionName(undefinedPerm); got != want {
		t.Errorf("GetPermissionName(undefined) = %q, want %q", got, want)
	}
}

func TestGetPermissionsText_AdminPermission(t *testing.T) {
	got := constants.GetPermissionsText(int64(constants.ADMIN))
	want := []string{"システム管理者（全権限）"}

	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("GetPermissionsText(ADMIN) = %v, want %v", got, want)
	}
}

func TestGetPermissionsText_SinglePermission(t *testing.T) {
	tests := []struct {
		name      string
		bitmask   int64
		wantTexts []string
	}{
		{
			name:      "USER_READ only",
			bitmask:   int64(constants.USER_READ),
			wantTexts: []string{"ユーザー読み取り"},
		},
		{
			name:      "APP_UPDATE only",
			bitmask:   int64(constants.APP_UPDATE),
			wantTexts: []string{"アプリ更新"},
		},
		{
			name:      "TOKEN_REVOKE only",
			bitmask:   int64(constants.TOKEN_REVOKE),
			wantTexts: []string{"トークン取り消し"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constants.GetPermissionsText(tt.bitmask)
			if len(got) != len(tt.wantTexts) {
				t.Errorf("GetPermissionsText(%d) returned %d permissions, want %d", tt.bitmask, len(got), len(tt.wantTexts))
				return
			}

			for _, want := range tt.wantTexts {
				if !slices.Contains(got, want) {
					t.Errorf("GetPermissionsText(%d) = %v, missing %q", tt.bitmask, got, want)
				}
			}
		})
	}
}

func TestGetPermissionsText_MultiplePermissions(t *testing.T) {
	tests := []struct {
		name      string
		bitmask   int64
		wantTexts []string
	}{
		{
			name:      "USER_READ and USER_CREATE",
			bitmask:   int64(constants.USER_READ | constants.USER_CREATE),
			wantTexts: []string{"ユーザー読み取り", "ユーザー作成"},
		},
		{
			name:      "all user management permissions",
			bitmask:   int64(constants.USER_READ | constants.USER_CREATE | constants.USER_UPDATE | constants.USER_DELETE | constants.USER_DISABLE),
			wantTexts: []string{"ユーザー読み取り", "ユーザー作成", "ユーザー更新", "ユーザー削除", "ユーザー無効化"},
		},
		{
			name:      "cross-category permissions",
			bitmask:   int64(constants.USER_READ | constants.APP_UPDATE | constants.TOKEN_REVOKE),
			wantTexts: []string{"ユーザー読み取り", "アプリ更新", "トークン取り消し"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := constants.GetPermissionsText(tt.bitmask)
			if len(got) != len(tt.wantTexts) {
				t.Errorf("GetPermissionsText(%d) returned %d permissions, want %d", tt.bitmask, len(got), len(tt.wantTexts))
				return
			}

			for _, want := range tt.wantTexts {
				if !slices.Contains(got, want) {
					t.Errorf("GetPermissionsText(%d) = %v, missing %q", tt.bitmask, got, want)
				}
			}
		})
	}
}

func TestGetPermissionsText_ZeroBitmask(t *testing.T) {
	got := constants.GetPermissionsText(0)

	if len(got) != 0 {
		t.Errorf("GetPermissionsText(0) = %v, want empty slice", got)
	}
}

func TestGetPermissionsText_AllAnnouncements(t *testing.T) {
	bitmask := int64(constants.ANNOUNCEMENT_CREATE | constants.ANNOUNCEMENT_UPDATE | constants.ANNOUNCEMENT_DELETE | constants.ANNOUNCEMENT_PIN)
	got := constants.GetPermissionsText(bitmask)
	wantTexts := []string{"お知らせ作成", "お知らせ編集", "お知らせ削除", "お知らせピン留め"}

	if len(got) != len(wantTexts) {
		t.Errorf("GetPermissionsText() returned %d permissions, want %d", len(got), len(wantTexts))
		return
	}

	for _, want := range wantTexts {
		if !slices.Contains(got, want) {
			t.Errorf("GetPermissionsText() = %v, missing %q", got, want)
		}
	}
}

func TestGetPermissionsText_AllRBACPermissions(t *testing.T) {
	bitmask := int64(constants.ROLE_MANAGE | constants.PERMISSION_MANAGE | constants.SESSION_MANAGE | constants.MFA_MANAGE)
	got := constants.GetPermissionsText(bitmask)
	wantTexts := []string{"ロール管理", "権限管理", "セッション管理", "多要素認証管理"}

	if len(got) != len(wantTexts) {
		t.Errorf("GetPermissionsText() returned %d permissions, want %d", len(got), len(wantTexts))
		return
	}

	for _, want := range wantTexts {
		if !slices.Contains(got, want) {
			t.Errorf("GetPermissionsText() = %v, missing %q", got, want)
		}
	}
}

func TestPermissionConsistency(t *testing.T) {
	// Verify that all permissions in PermissionNames map have proper names
	for perm, name := range constants.PermissionNames {
		if name == "" {
			t.Errorf("Permission %v has empty name", perm)
		}
	}

	// Verify that GetPermissionName matches PermissionNames for defined permissions
	for perm, expectedName := range constants.PermissionNames {
		got := constants.GetPermissionName(perm)
		if got != expectedName {
			t.Errorf("GetPermissionName(%v) = %q, want %q", perm, got, expectedName)
		}
	}
}

func TestAliasPermissions(t *testing.T) {
	// Verify backward compatibility aliases
	if constants.USER_WRITE != constants.USER_UPDATE {
		t.Errorf("USER_WRITE should equal USER_UPDATE")
	}
	if constants.USER_APPROVE != constants.USER_UPDATE {
		t.Errorf("USER_APPROVE should equal USER_UPDATE")
	}
	if constants.PROFILE_READ != constants.USER_READ {
		t.Errorf("PROFILE_READ should equal USER_READ")
	}
	if constants.PROFILE_WRITE != constants.USER_UPDATE {
		t.Errorf("PROFILE_WRITE should equal USER_UPDATE")
	}
	if constants.EXTERNAL_IDENTITY_READ != constants.USER_READ {
		t.Errorf("EXTERNAL_IDENTITY_READ should equal USER_READ")
	}
	if constants.EXTERNAL_IDENTITY_WRITE != constants.USER_UPDATE {
		t.Errorf("EXTERNAL_IDENTITY_WRITE should equal USER_UPDATE")
	}
	if constants.EXTERNAL_IDENTITY_DELETE != constants.USER_UPDATE {
		t.Errorf("EXTERNAL_IDENTITY_DELETE should equal USER_UPDATE")
	}
	if constants.AUDIT_LOG_READ != constants.AUDIT_READ {
		t.Errorf("AUDIT_LOG_READ should equal AUDIT_READ")
	}
}
