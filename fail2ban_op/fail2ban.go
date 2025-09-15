package fail2ban_op

import (
	"context"
	"errors"
	"time"
)

type Fail2ban struct {
	LoginErrorWindow    int // 登录错误时间窗口（秒）
	LoginErrorThreshold int // 登录错误次数阈值
	BanDuration         int // 封禁时长（秒）

	logCh    chan logIP
	cancelFn context.CancelFunc

	// 记录每个 IP 的登录错误时间列表
	ipErrors map[string][]time.Time
	// 记录被封禁的 IP 及解封时间
	bannedIPs map[string]time.Time
}

type logIP struct {
	t     time.Time
	ipstr string
}

func (ban *Fail2ban) startLogRead(ctx context.Context) {
	go FollowLogRead(ctx, func(entry LogEntry) {
		t, err := time.Parse("Mon Jan 2 15:04:05 2006", entry.Time)
		if err != nil {
			t = time.Now()
		}
		if ipstr, ok := IsDropbearBadPasswordLog(&entry); ok {
			ban.logCh <- logIP{t: t, ipstr: ipstr}
		} else if ipstr, ok := IsUhttpdLoginErrorLog(&entry); ok {
			ban.logCh <- logIP{t: t, ipstr: ipstr}
		}
	})
}

func (ban *Fail2ban) run(ctx context.Context) {
	for {
		if err := ban.runOnce(ctx); err != nil {
			return
		}
	}
}

func (ban *Fail2ban) runOnce(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case lip := <-ban.logCh:
		return ban.logNewIP(lip)
	}
}

func (ban *Fail2ban) logNewIP(lip logIP) error {
	if ban.ipErrors == nil {
		ban.ipErrors = make(map[string][]time.Time)
	}
	if ban.bannedIPs == nil {
		ban.bannedIPs = make(map[string]time.Time)
	}

	// 检查 IP 是否已被封禁
	if unbanTime, banned := ban.bannedIPs[lip.ipstr]; banned {
		if time.Now().Before(unbanTime) {
			return nil // 已封禁，忽略
		} else {
			// 解封，清理数据
			delete(ban.bannedIPs, lip.ipstr)
			delete(ban.ipErrors, lip.ipstr)
		}
	}

	// 记录错误时间
	times := ban.ipErrors[lip.ipstr]
	times = append(times, lip.t)
	windowStart := lip.t.Add(-time.Duration(ban.LoginErrorWindow) * time.Second)
	// 清理窗口外的时间
	var newTimes []time.Time
	for _, t := range times {
		if t.After(windowStart) {
			newTimes = append(newTimes, t)
		}
	}
	ban.ipErrors[lip.ipstr] = newTimes

	// 检查是否超过阈值
	if len(newTimes) >= ban.LoginErrorThreshold {
		ban.bannedIPs[lip.ipstr] = time.Now().Add(time.Duration(ban.BanDuration) * time.Second)
		return nil // 已封禁
	}
	return nil
}

func (ban *Fail2ban) Start(ctx context.Context) error {
	if ban.cancelFn != nil {
		return errors.New("already started")
	}
	newCtx, cancelFn := context.WithCancel(ctx)
	ban.cancelFn = cancelFn
	ban.startLogRead(newCtx)
	go ban.run(newCtx)
	return nil
}

func (ban *Fail2ban) Stop() error {
	ban.cancelFn()
	return nil
}
