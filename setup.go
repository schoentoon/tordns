package tordns

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/caddyserver/caddy"
	"github.com/cretz/bine/control"
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

	c.OnStartup(func() error {
		return t.setupConnection()
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		t.Next = next
		return t
	})

	return nil
}

func configParse(c *caddy.Controller) (TorDnsPlugin, error) {
	t := TorDnsPlugin{
		TorDns: &TorDns{
			addrCallback: make(chan control.Event),
			pool:         []chan control.AddrMapEvent{},
		},
	}

	if c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "controlsocket":
				if !c.NextArg() {
					return t, c.ArgErr()
				}
				t.controlSocketPath = c.Val()
			}
		}
	}

	return t, nil
}
