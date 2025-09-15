package fail2ban_op

import (
	"regexp"
)

var uhttpdLoginErrorRegex = regexp.MustCompile(`daemon\.err uhttpd\[\d+\]: \[info\] luci: failed login on / for root from ([0-9.]+)`)

// IsUhttpdLoginErrorLog 判断日志行是否为 uhttpd 的 HTTP 密码错误日志，并提取 IP。
func IsUhttpdLoginErrorLog(entry *LogEntry) (ip string, ok bool) {
	matches := uhttpdLoginErrorRegex.FindStringSubmatch(entry.Text)
	if len(matches) == 2 {
		ip = matches[1]
		ok = true
		return
	}
	return "", false
}
