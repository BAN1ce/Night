package utils

/**
topic中是否包含 +，# 通配符
*/
func HasWildcard(topicSlice []string) bool {

	for _, v := range topicSlice {
		if v == "#" || v == "+" {
			return true
		}
	}
	return false
}
