package fail2ban_op

import (
	"testing"
)

func TestIsDropbearBadPasswordLog(t *testing.T) {
	tests := []struct {
		name    string
		entry   LogEntry
		wantIP  string
		wantMat bool
	}{
		{
			name:    "Bad password with angle brackets",
			entry:   LogEntry{Text: "dropbear[1234]: Bad password attempt for 'root' from <192.168.1.100:12345>"},
			wantIP:  "192.168.1.100",
			wantMat: true,
		},
		{
			name:    "Bad password without angle brackets",
			entry:   LogEntry{Text: "dropbear[1234]: Bad password attempt for 'root' from 10.0.0.2:2222"},
			wantIP:  "10.0.0.2",
			wantMat: true,
		},
		{
			name:    "Exit before auth",
			entry:   LogEntry{Text: "dropbear[1234]: Exit before auth from <172.16.0.5:2022>"},
			wantIP:  "172.16.0.5",
			wantMat: true,
		},
		{
			name:    "Non-dropbear log",
			entry:   LogEntry{Text: "sshd[1234]: Bad password attempt for 'root' from <192.168.1.100:12345>"},
			wantIP:  "",
			wantMat: false,
		},
		{
			name:    "No IP in log",
			entry:   LogEntry{Text: "dropbear[1234]: Bad password attempt for 'root'"},
			wantIP:  "",
			wantMat: false,
		},
		{
			name:    "ipv6 normal",
			entry:   LogEntry{Text: "authpriv.warn dropbear[17841]: Bad password attempt for 'root' from fdac:b153:8eb3:1:be24:11ff:feb9:9fa1:51422"},
			wantIP:  "fdac:b153:8eb3:1:be24:11ff:feb9:9fa1",
			wantMat: true,
		},
		{
			name:    "ipv6 end",
			entry:   LogEntry{Text: "authpriv.info dropbear[17841]: Exit before auth from <fdac:b153:8eb3:1:be24:11ff:feb9:9fa1:51422>: (user 'root', 3 fails): Max auth tries reached - user 'root'"},
			wantIP:  "fdac:b153:8eb3:1:be24:11ff:feb9:9fa1",
			wantMat: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, matched := IsDropbearBadPasswordLog(&tt.entry)
			if ip != tt.wantIP || matched != tt.wantMat {
				t.Errorf("got (%q, %v), want (%q, %v)", ip, matched, tt.wantIP, tt.wantMat)
			}
		})
	}
}
