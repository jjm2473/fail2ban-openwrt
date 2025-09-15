package fail2ban_op

import (
	"testing"
)

func TestIsUhttpdLoginErrorLog(t *testing.T) {
	cases := []struct {
		name   string
		log    string
		wantIP string
		wantOK bool
	}{
		{
			name:   "match",
			log:    "daemon.err uhttpd[1234]: [info] luci: failed login on / for root from 192.168.1.100",
			wantIP: "192.168.1.100",
			wantOK: true,
		},
		{
			name:   "not match",
			log:    "daemon.info uhttpd[1234]: [info] luci: successful login on / for root from 192.168.1.100",
			wantIP: "",
			wantOK: false,
		},
		{
			name:   "wrong format",
			log:    "random log line",
			wantIP: "",
			wantOK: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entry := &LogEntry{Text: c.log}
			ip, ok := IsUhttpdLoginErrorLog(entry)
			if ip != c.wantIP || ok != c.wantOK {
				t.Errorf("got (%q, %v), want (%q, %v)", ip, ok, c.wantIP, c.wantOK)
			}
		})
	}
}
