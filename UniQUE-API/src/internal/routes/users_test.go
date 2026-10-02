package routes

import "testing"

func TestShouldRequestExternalEmailChange(t *testing.T) {
	tests := []struct {
		name           string
		currentEmail   string
		pendingEmail   string
		requestedEmail string
		want           bool
	}{
		{
			name:           "new address requires verification",
			currentEmail:   "current@example.com",
			requestedEmail: "new@example.com",
			want:           true,
		},
		{
			name:           "current address is unchanged",
			currentEmail:   "current@example.com",
			requestedEmail: "current@example.com",
			want:           false,
		},
		{
			name:           "pending address is unchanged",
			currentEmail:   "current@example.com",
			pendingEmail:   "pending@example.com",
			requestedEmail: "pending@example.com",
			want:           false,
		},
		{
			name:           "different address replaces pending verification",
			currentEmail:   "current@example.com",
			pendingEmail:   "pending@example.com",
			requestedEmail: "other@example.com",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRequestExternalEmailChange(tt.currentEmail, tt.pendingEmail, tt.requestedEmail)
			if got != tt.want {
				t.Fatalf("shouldRequestExternalEmailChange() = %v, want %v", got, tt.want)
			}
		})
	}
}
