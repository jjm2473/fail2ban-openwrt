package fail2ban_op

import (
	"fmt"
	"os/exec"
)

// AddIPToIPSet adds the given IP to the specified ipset or nft set.
func AddIPToIPSet(ip string, setName string) error {
	// Try ipset first
	cmd := exec.Command("ipset", "add", setName, ip)
	if err := cmd.Run(); err == nil {
		return nil
	}
	// If ipset fails, try nft
	// nft add element ip filter <set> { <ip> }
	nftCmd := exec.Command("nft", "add", "element", "ip", "filter", setName, fmt.Sprintf("{ %s }", ip))
	if err := nftCmd.Run(); err != nil {
		return err
	}
	return nil
}
