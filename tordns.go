package tordns

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/cretz/bine/control"
	"github.com/miekg/dns"
)

type TorDnsPlugin struct {
	*TorDns

	Next plugin.Handler
}

type TorDns struct {
	Conn *control.Conn

	controlSocketPath string
	addrCallback      chan control.Event

	poolMutex sync.RWMutex
	pool      []chan control.AddrMapEvent
}

func (t *TorDns) init() error {
	if t.Conn == nil {
		return fmt.Errorf("conn shouldn't be nil anymore!")
	}

	err := t.Conn.Authenticate("")
	if err != nil {
		return err
	}

	go t.consumeAddrCallbacks()

	return t.Conn.AddEventListener(t.addrCallback, control.EventCodeAddrMap)
}

func (t *TorDns) consumeAddrCallbacks() {
	go t.Conn.HandleEvents(context.Background())
	for event := range t.addrCallback {
		addr, ok := event.(*control.AddrMapEvent)
		if !ok {
			log.Errorf("we somehow got an event we don't want? %#v", event)
		}

		t.poolMutex.RLock()
		for _, ch := range t.pool {
			ch <- *addr
		}
		t.poolMutex.RUnlock()
	}
}

func (t *TorDns) newConsumer() chan control.AddrMapEvent {
	out := make(chan control.AddrMapEvent)

	t.poolMutex.Lock()
	defer t.poolMutex.Unlock()

	t.pool = append(t.pool, out)

	return out
}

func (t *TorDns) unregister(del chan control.AddrMapEvent) {
	t.poolMutex.Lock()
	defer t.poolMutex.Unlock()

	for idx, ch := range t.pool {
		if ch == del {
			t.pool[len(t.pool)-1], t.pool[idx] = t.pool[idx], t.pool[len(t.pool)-1]
			t.pool = t.pool[:len(t.pool)-1]
		}
	}
}

func (t TorDnsPlugin) Name() string {
	return "tordns"
}

func (t TorDnsPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	var err error

	ch := t.newConsumer()
	defer t.unregister(ch)

	switch state.QType() {
	case dns.TypeA:
		fallthrough
	case dns.TypeAAAA:
		err = t.Conn.ResolveAsync(qname, false)
	default:
		return plugin.NextOrFailure(t.Name(), t.Next, ctx, w, r)
	}

	if err != nil {
		return dns.RcodeServerFailure, err
	}

	for {
		select {
		case <-ctx.Done():
			return plugin.NextOrFailure(t.Name(), t.Next, ctx, w, r)
		case answer := <-ch:
			if answer.Address == qname {
				answers := []dns.RR{}
				until := time.Until(answer.Expires).Seconds()
				if until < 0 {
					until = 0
				}
				ttl := uint32(until)

				switch state.QType() {
				case dns.TypeA:
					answers = a(qname, ttl, net.ParseIP(answer.NewAddress))
				case dns.TypeAAAA:
					answers = aaaa(qname, ttl, net.ParseIP(answer.NewAddress))
				}

				m := new(dns.Msg)
				m.SetReply(r)
				m.Authoritative = true
				m.Answer = answers

				w.WriteMsg(m)
				return dns.RcodeSuccess, nil
			}
		}
	}
}

func a(zone string, ttl uint32, ip net.IP) []dns.RR {
	answers := make([]dns.RR, 1)

	r := new(dns.A)
	r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
	r.A = ip

	answers[0] = r

	return answers
}

func aaaa(zone string, ttl uint32, ip net.IP) []dns.RR {
	answers := make([]dns.RR, 1)
	r := new(dns.AAAA)
	r.Hdr = dns.RR_Header{Name: zone, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
	r.AAAA = ip

	answers[0] = r

	return answers
}
