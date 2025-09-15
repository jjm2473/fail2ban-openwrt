package main

import (
	"fmt"
	"time"

	"github.com/linkease/fail2ban_op/fail2ban_op"
)

func main() {
	timeLayout := "Mon Jan 2 15:04:05 2006"

	fail2ban_op.FollowLogRead(func(entry fail2ban_op.LogEntry) {
		t, err := time.Parse(timeLayout, entry.Time)
		if err != nil {
			fmt.Println("parse time error:", err, "entry.Time=", entry.Time)
			return
		}
		fmt.Println("entry time=", t, "text=", entry.Text)
		ipstr, matched := fail2ban_op.IsDropbearBadPasswordLog(&entry)
		if !matched {
			ipstr, matched = fail2ban_op.IsUhttpdLoginErrorLog(&entry)
		}
		fmt.Println("ipstr=", ipstr, "matched=", matched)
	})
}
