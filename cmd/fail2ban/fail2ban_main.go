package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/linkease/fail2ban_op/api"
	"github.com/linkease/fail2ban_op/fail2ban_op"

	"github.com/urfave/cli/v2"
)

var (
	BuildVersion string
	BuildDate    string
)

// BlockIP 实现，调用 block.go 的 AddIPToIPSet 和 RemoveIPFromIPSet

type ipsetBlocker struct{}

func (b *ipsetBlocker) BlockIP(ipstr string) error {
	return fail2ban_op.AddIPToIPSet(ipstr)
}

func (b *ipsetBlocker) UnblockIP(ipstr string) error {
	return fail2ban_op.RemoveIPFromIPSet(ipstr)
}

func main() {
	app := &cli.App{
		Name:  "fail2ban-openwrt",
		Usage: "Fail2ban for OpenWrt",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "window",
				Usage: "Login error window (seconds)",
				Value: 600,
			},
			&cli.IntFlag{
				Name:  "threshold",
				Usage: "Login error threshold",
				Value: 10,
			},
			&cli.IntFlag{
				Name:  "ban-duration",
				Usage: "Ban duration (minutes)",
				Value: 60 * 24,
			},
			&cli.BoolFlag{
				Name:  "show-banned-ips",
				Usage: "Show currently banned IPs, for debugging",
				Value: false,
			},
		},
		Action: func(c *cli.Context) error {
			return mainAction(c)
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Show the current version",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "more",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					return versionAction(c)
				},
			},
			{
				Name:  "show-ipset",
				Usage: "Show the ipset used by fail2ban-openwrt",
				Action: func(c *cli.Context) error {
					return showIPsetsAction()
				},
			},
			{
				Name:  "remove-ipset",
				Usage: "Remove the ipset used by fail2ban-openwrt",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "ip",
					},
				},
				Action: func(c *cli.Context) error {
					return removeIPsetAction(c)
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	window := c.Int("window")
	threshold := c.Int("threshold")
	banDuration := c.Int("ban-duration") * 60
	if c.Bool("show-banned-ips") {
		fail2ban_op.ShowBannedIPs = true
	}

	// 重启清空现有的所有记录
	fail2ban_op.FlushIPSet()

	blocker := &ipsetBlocker{}
	fail2ban := fail2ban_op.NewFail2ban(window, threshold, banDuration, blocker)
	if err := fail2ban.Start(context.Background()); err != nil {
		log.Fatalf("fail2ban start error: %v", err)
	}
	defer fail2ban.Stop()

	fmt.Printf("fail2ban started. window=%d threshold=%d banDuration=%d\n", window, threshold, banDuration)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("Received stop signal, stopping fail2ban...")
	return nil
}

func versionAction(c *cli.Context) error {
	if c.Bool("more") {
		fmt.Println(api.VERSION, BuildVersion, BuildDate)
	} else {
		fmt.Println(api.VERSION)
	}
	return nil
}

func removeIPsetAction(c *cli.Context) error {
	ip := c.String("ip")
	if ip == "" {
		return fmt.Errorf("IP address is required")
	}
	return fail2ban_op.RemoveIPFromIPSet(ip)
}

func showIPsetsAction() error {
	ipCh := fail2ban_op.ShowIPsets()
	for ip := range ipCh {
		fmt.Println(ip)
	}
	return nil
}
