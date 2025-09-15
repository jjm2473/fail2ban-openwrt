package fail2ban_op

import (
	"testing"
	"time"
)

// mockBlockIP implements BlockIP for testing
type mockBlockIP struct {
	blocked   map[string]bool
	unblocked map[string]bool
}

func (m *mockBlockIP) BlockIP(ipstr string) error {
	m.blocked[ipstr] = true
	return nil
}
func (m *mockBlockIP) UnblockIP(ipstr string) error {
	m.unblocked[ipstr] = true
	return nil
}

func TestLogNewIP_BanThreshold(t *testing.T) {
	mock := &mockBlockIP{blocked: make(map[string]bool), unblocked: make(map[string]bool)}
	ban := NewFail2ban(10, 3, 60, mock)
	ip := "1.2.3.4"
	t0 := time.Now()
	for i := 0; i < 3; i++ {
		lip := logIP{t: t0.Add(time.Duration(i) * time.Second), ipstr: ip}
		ban.logNewIP(lip)
	}
	if !mock.blocked[ip] {
		t.Errorf("IP should be blocked after threshold reached")
	}
	if _, ok := ban.bannedIPs[ip]; !ok {
		t.Errorf("IP should be in bannedIPs map")
	}
}

func TestLogNewIP_Window(t *testing.T) {
	mock := &mockBlockIP{blocked: make(map[string]bool), unblocked: make(map[string]bool)}
	ban := NewFail2ban(2, 3, 60, mock)
	ip := "5.6.7.8"
	t0 := time.Now()
	ban.logNewIP(logIP{t: t0, ipstr: ip})
	ban.logNewIP(logIP{t: t0.Add(1 * time.Second), ipstr: ip})
	ban.logNewIP(logIP{t: t0.Add(3 * time.Second), ipstr: ip}) // outside window
	if mock.blocked[ip] {
		t.Errorf("IP should not be blocked if errors are outside window")
	}
}

func TestLogNewIP_AlreadyBanned(t *testing.T) {
	mock := &mockBlockIP{blocked: make(map[string]bool), unblocked: make(map[string]bool)}
	ban := NewFail2ban(10, 2, 60, mock)
	ip := "9.8.7.6"
	ban.bannedIPs[ip] = &banIP{ipstr: ip, bannedTime: time.Now().Add(60 * time.Second), el: ban.bannedTimeoutList.PushBack(nil)}
	oldTime := ban.bannedIPs[ip].bannedTime
	ban.logNewIP(logIP{t: time.Now(), ipstr: ip})
	if ban.bannedIPs[ip].bannedTime == oldTime {
		t.Errorf("Banned time should be updated for already banned IP")
	}
}
