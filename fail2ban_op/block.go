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
		nftCmd := exec.Command("nft", "list", "set", "inet", "fw4", ipsetName)
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
