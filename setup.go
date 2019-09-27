package tordns

import (
	"errors"
	"net"
	"strconv"

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
		TorDns: &TorDns{
			maxRetries:  3,
			maxPoolSize: 10,
			connPool:    make(chan net.Conn, 10),
		},
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
				if !c.NextArg() {
					return t, c.ArgErr()
				}
				dialer, err := proxy.SOCKS5("tcp", c.Val(), nil, proxy.Direct)
				if err != nil {
					return t, err
				}
				t.Proxy = dialer
			case "retries":
				if !c.NextArg() {
					return t, c.ArgErr()
				}
				retries, err := strconv.Atoi(c.Val())
				if err != nil {
					return t, err
				}
				t.maxRetries = retries
			case "poolsize":
				if !c.NextArg() {
					return t, c.ArgErr()
				}
				poolsize, err := strconv.Atoi(c.Val())
				if err != nil {
					return t, err
				}
				t.maxPoolSize = int32(poolsize)
				t.connPool = make(chan net.Conn, poolsize)
			}
		}
	}

	return t, nil
}
