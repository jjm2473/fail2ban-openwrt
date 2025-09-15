package fail2ban_op

import (
	"container/list"
	"context"
	"errors"
	"time"
)

const BlockIPMax = 1024 * 1024 // 最大封禁 IP 数量

type BlockIP interface {
	BlockIP(ipstr string) error
	UnblockIP(ipstr string) error
}

type Fail2ban struct {
	LoginErrorWindow    int // 登录错误时间窗口（秒）
	LoginErrorThreshold int // 登录错误次数阈值
	BanDuration         int // 封禁时长（秒）
	blockIPMax          int // 最大封禁 IP 数量
	block               BlockIP

	logCh    chan logIP
	cancelFn context.CancelFunc

	// 记录每个 IP 的登录错误时间列表
	ipErrors map[string][]time.Time
	// 记录被封禁的 IP 及解封时间
	bannedIPs         map[string]*banIP
	bannedTimeoutList *list.List
}

type logIP struct {
	t     time.Time
	ipstr string
}

type banIP struct {
	el         *list.Element
	bannedTime time.Time
	ipstr      string
}

func NewFail2ban(loginErrorWindow, loginErrorThreshold, banDuration int,
	block BlockIP) *Fail2ban {
	return &Fail2ban{
		LoginErrorWindow:    loginErrorWindow,
		LoginErrorThreshold: loginErrorThreshold,
		BanDuration:         banDuration,
		blockIPMax:          BlockIPMax,
		block:               block,
		ipErrors:            make(map[string][]time.Time),
		bannedIPs:           make(map[string]*banIP),
		bannedTimeoutList:   &list.List{},
		logCh:               make(chan logIP, 100),
	}
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
	tick := time.NewTicker(time.Second * time.Duration(ban.BanDuration))
	for {
		if err := ban.runOnce(ctx, tick); err != nil {
			return
		}
	}
}

func (ban *Fail2ban) runOnce(ctx context.Context, tick *time.Ticker) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case lip := <-ban.logCh:
		return ban.logNewIP(lip)
	case <-tick.C:
		return ban.checkTimeouts()
	}
}

func (ban *Fail2ban) logNewIP(lip logIP) error {
	// 检查 IP 是否已被封禁
	if bip, banned := ban.bannedIPs[lip.ipstr]; banned {
		bip.bannedTime = time.Now().Add(time.Duration(ban.BanDuration) * time.Second)
		ban.bannedTimeoutList.MoveToBack(bip.el)
		return nil
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
		bip := &banIP{
			ipstr:      lip.ipstr,
			bannedTime: time.Now().Add(time.Duration(ban.BanDuration) * time.Second),
		}
		bip.el = ban.bannedTimeoutList.PushBack(bip)
		ban.bannedIPs[lip.ipstr] = bip
		ban.block.BlockIP(lip.ipstr)

		if len(ban.bannedIPs) > ban.blockIPMax {
			// 删除最早封禁的 IP，省得内存使用过大
			// 如果别人使用超级多的 IP 攻击，内存会被耗尽，则也没办法防护了
			first := ban.bannedTimeoutList.Front()
			fbip := first.Value.(*banIP)
			delete(ban.bannedIPs, fbip.ipstr)
			ban.bannedTimeoutList.Remove(first)
			ban.block.UnblockIP(fbip.ipstr)
		}

		return nil
	}
	return nil
}

func (ban *Fail2ban) checkTimeouts() error {
	if ban.bannedTimeoutList.Len() == 0 {
		return nil
	}
	now := time.Now()
	for e := ban.bannedTimeoutList.Front(); e != nil; {
		bip := e.Value.(*banIP)
		if now.Before(bip.bannedTime) {
			// 还未到解封时间，后续也不需要检查了
			break
		}
		// 超时，解封
		delete(ban.bannedIPs, bip.ipstr)
		ban.bannedTimeoutList.Remove(e)
		ban.block.UnblockIP(bip.ipstr)
		e = ban.bannedTimeoutList.Front() // 重新从头遍历
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
