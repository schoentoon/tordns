package tordns

import (
	"errors"
	"fmt"

	"golang.org/x/net/proxy"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/caddyserver/caddy"
)

var log = clog.NewWithPlugin("tordns")

func init() {
	caddy.RegisterPlugin("tordns", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	t, err := configParse(c)
	if err != nil {
		return plugin.Error("tordns", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		t.Next = next
		return t
	})

	return nil
}

func configParse(c *caddy.Controller) (TorDnsPlugin, error) {
	t := TorDnsPlugin{
		TorDns: &TorDns{},
	}

	if c.Next() {
		args := c.RemainingArgs()
		if len(args) < 1 {
			return t, errors.New("no hidden services specified?")
		}
		t.TorDns.hiddenService = args[0]
		for c.NextBlock() {
			switch c.Val() {
			case "proxy":
				args = c.RemainingArgs()
				if len(args) != 1 {
					return t, fmt.Errorf("Expected just one argument for tor address, got %d", len(args))
				}
				dialer, err := proxy.SOCKS5("tcp", args[0], nil, proxy.Direct)
				if err != nil {
					return t, err
				}
				t.Proxy = dialer
			}
		}
	}

	return t, nil
}
