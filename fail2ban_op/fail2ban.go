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
	stopCh   chan struct{}
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
