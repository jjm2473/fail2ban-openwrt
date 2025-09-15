package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/linkease/fail2ban_op/fail2ban_op"

	"github.com/urfave/cli/v2"
)

// BlockIP 实现，调用 block.go 的 AddIPToIPSet 和 RemoveIPFromIPSet

type ipsetBlocker struct{}

func (b *ipsetBlocker) BlockIP(ipstr string) error {
	return fail2ban_op.AddIPToIPSet([]string{ipstr})
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
				Value: 60,
			},
			&cli.IntFlag{
				Name:  "threshold",
				Usage: "Login error threshold",
				Value: 5,
			},
			&cli.IntFlag{
				Name:  "ban-duration",
				Usage: "Ban duration (seconds)",
				Value: 600,
			},
		},
		Action: func(c *cli.Context) error {
			window := c.Int("window")
			threshold := c.Int("threshold")
			banDuration := c.Int("ban-duration")

			blocker := &ipsetBlocker{}
			fail2ban := fail2ban_op.NewFail2ban(window, threshold, banDuration, blocker)
			if err := fail2ban.Start(context.Background()); err != nil {
				log.Fatalf("fail2ban start error: %v", err)
			}
			defer fail2ban.Stop()

			log.Printf("fail2ban started. window=%d threshold=%d banDuration=%d", window, threshold, banDuration)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			log.Println("Received stop signal, stopping fail2ban...")
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
