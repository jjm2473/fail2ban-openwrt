package fail2ban_op

import (
	"fmt"
	"os/exec"
	"strings"
)

const ipsetName = "fail2banop"

func AddIPToIPSet(ips []string) error {
	if _, err := exec.LookPath("nft"); err == nil {
		ips := strings.Join(ips, ", ")
		nftCmd := exec.Command("nft", "add", "element", "inet", "fw4", ipsetName, fmt.Sprintf("{ %s }", ips))
		if err := nftCmd.Run(); err != nil {
			return err
		}
		return nil
	}
	// Fallback to ipset
	for _, ip := range ips {
		cmd := exec.Command("ipset", "add", ipsetName, ip)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func FlushIPSet() error {
	if _, err := exec.LookPath("nft"); err == nil {
		nftCmd := exec.Command("nft", "flush", "set", "inet", "fw4", ipsetName)
		if err := nftCmd.Run(); err != nil {
			return err
		}
		return nil
	}
	// Fallback to ipset flush
	cmd := exec.Command("ipset", "flush", ipsetName)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func RemoveIPFromIPSet(ip string) error {
	if _, err := exec.LookPath("nft"); err == nil {
		nftCmd := exec.Command("nft", "delete", "element", "inet", "fw4", ipsetName, fmt.Sprintf("{ %s }", ip))
		if err := nftCmd.Run(); err != nil {
			return err
		}
		return nil
	}
	cmd := exec.Command("ipset", "del", ipsetName, ip)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func ShowIPsets() chan string {
	ips := make(chan string, 16)
	go func() {
		defer close(ips)

		// Try nft first
		if _, err := exec.LookPath("nft"); err == nil {
			cmd := exec.Command("nft", "list", "set", "inet", "fw4", ipsetName)
			out, err := cmd.Output()
			if err == nil {
				lines := strings.Split(string(out), "\n")
				found := false
				var ipsBuffer []string
				for _, line := range lines {
					if strings.Contains(line, "elements = {") {
						found = true
						line = strings.TrimPrefix(line, "elements = {")
						ipsBuffer = append(ipsBuffer, strings.TrimSpace(line))
					} else if found && strings.Contains(line, "}") {
						line = strings.TrimSuffix(line, "}")
						ipsBuffer = append(ipsBuffer, strings.TrimSpace(line))
						break
					} else if found {
						ipsBuffer = append(ipsBuffer, strings.TrimSpace(line))
					}
				}
				if found {
					for _, ip := range strings.Split(strings.Join(ipsBuffer, ""), ",") {
						ip = strings.TrimSpace(ip)
						if ip != "" {
							ips <- ip
						}
					}
					return
				}
			}
		}

		// Fallback to ipset
		cmd := exec.Command("ipset", "list", ipsetName)
		out, err := cmd.Output()
		if err != nil {
			return
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, " ") {
				ip := strings.Fields(line)[0]
				ips <- ip
			}
		}
	}()
	return ips
}
