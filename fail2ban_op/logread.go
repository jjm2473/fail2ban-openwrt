package fail2ban_op

import (
	"bufio"
	"context"
	"os/exec"
	"regexp"
	"strings"
)

// LogEntry 表示一条日志记录，包含时间和文本内容
type LogEntry struct {
	Time string
	Text string
}

// FollowLogRead 持续跟踪 logread -f 输出，解析时间和文本
func FollowLogRead(ctx context.Context, handleEntry func(LogEntry)) error {
	cmd := exec.CommandContext(ctx, "logread", "-f")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	// 日志时间格式: Mon Sep 15 15:57:09 2025
	timeRe := regexp.MustCompile(`^[A-Z][a-z]{2} [A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2} \d{4}`)
	for scanner.Scan() {
		line := scanner.Text()
		timeStr := ""
		text := line
		if m := timeRe.FindString(line); m != "" {
			timeStr = m
			text = strings.TrimSpace(line[len(m):])
		}
		handleEntry(LogEntry{Time: timeStr, Text: text})
	}
	return scanner.Err()
}
