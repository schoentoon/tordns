package tordns

import (
	"context"

	"golang.org/x/net/proxy"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

type TorDnsPlugin struct {
	*TorDns

	Next plugin.Handler
}

type TorDns struct {
	hiddenService string

	Proxy proxy.Dialer
}

func (t TorDnsPlugin) Name() string {
	return "tordns"
}

func (t TorDnsPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	conn, err := t.Proxy.Dial("tcp", t.hiddenService)
	if err != nil {
		log.Debugf("connect error: %s\n", err)
		return dns.RcodeServerFailure, err
	}
	defer conn.Close()

	co := &dns.Conn{Conn: conn}
	defer co.Close()

	err = co.WriteMsg(r)
	if err != nil {
		log.Debugf("exchange error: %s\n", err)
		return dns.RcodeServerFailure, err
	}

	out, err  := co.ReadMsg()
	if err != nil {
		log.Debugf("exchange error: %s\n", err)
		return dns.RcodeServerFailure, err
	}
	w.WriteMsg(out)
	return dns.RcodeSuccess, nil
}