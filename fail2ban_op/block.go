package fail2ban_op

import (
	"fmt"
	"os/exec"
)

// AddIPToIPSet adds the given IP to the specified ipset or nft set.
func AddIPToIPSet(ip string, setName string) error {
	// Check if nft exists
	if _, err := exec.LookPath("nft"); err == nil {
		// nft add element ip filter <set> { <ip> }
		nftCmd := exec.Command("nft", "add", "element", "ip", "filter", setName, fmt.Sprintf("{ %s }", ip))
		if err := nftCmd.Run(); err != nil {
			return err
		}
		return nil
	}
	// Fallback to ipset
	cmd := exec.Command("ipset", "add", setName, ip)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
