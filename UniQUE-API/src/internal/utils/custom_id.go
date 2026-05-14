package utils

func IsValidCustomID(id string) bool {
	if len(id) < 3 || len(id) > 30 {
		return false
	}
	// 先頭と末尾は英数字でなければならない
	if !isAlphanumeric(rune(id[0])) || !isAlphanumeric(rune(id[len(id)-1])) {
		return false
	}
	// 文字は英数字、アンダースコア、ハイフンのみ許可
	for _, r := range id {
		if !isAlphanumeric(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
