package util

func AlphabeticScopeString(scope string) string {
	scopes := splitScope(scope)
	var result string
	for i, s := range scopes {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}

func MergeScopes(scopes1, scopes2 string) string {
	scopeMap := map[string]bool{}
	for _, s := range splitScope(scopes1) {
		scopeMap[s] = true
	}
	for _, s := range splitScope(scopes2) {
		scopeMap[s] = true
	}
	var result string
	first := true
	for s := range scopeMap {
		if !first {
			result += " "
		}
		result += s
		first = false
	}
	return result
}

func CompareScopes(scope1, scope2 string) bool {
	scopeMap1 := map[string]bool{}
	for _, s := range splitScope(scope1) {
		scopeMap1[s] = true
	}
	scopeMap2 := map[string]bool{}
	for _, s := range splitScope(scope2) {
		scopeMap2[s] = true
	}
	if len(scopeMap1) != len(scopeMap2) {
		return false
	}
	for s := range scopeMap1 {
		if !scopeMap2[s] {
			return false
		}
	}
	return true
}

func splitScope(scope string) []string {
	var result []string
	current := ""
	for _, c := range scope {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func ContainsScope(scopes, targetScope string) bool {
	scopeList := splitScope(scopes)
	for _, s := range scopeList {
		if s == targetScope {
			return true
		}
	}
	return false
}
