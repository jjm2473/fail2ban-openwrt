package fail2ban_op

import (
	"regexp"
	"strings"
)

// IsDropbearBadPasswordLog 判断日志是否为 dropbear 的 SSH 密码错误日志，并提取 IP
func IsDropbearBadPasswordLog(entry *LogEntry) (ip string, matched bool) {
	// 合并后的正则，匹配 from <IP:PORT> 或 from IP:PORT，支持 IPv4 和 IPv6
	pattern := `from <?((?:[0-9.]+|\[?[0-9a-fA-F:]+\]?)):\d+>?`
	re := regexp.MustCompile(pattern)
	if strings.Contains(entry.Text, "dropbear") &&
		(strings.Contains(entry.Text, "Bad password attempt") || strings.Contains(entry.Text, "Exit before auth from")) {
		matches := re.FindStringSubmatch(entry.Text)
		if len(matches) == 2 {
			return matches[1], true
		}
	}
	return "", false
}
