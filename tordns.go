package tordns

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

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

	maxRetries  int
	maxPoolSize int32
	poolSize    int32
	connPool    chan net.Conn

	Proxy proxy.Dialer
}

func (t *TorDns) GetConnection() net.Conn {
	select {
	case conn := <-t.connPool:
		return conn
	default:
		if atomic.LoadInt32(&t.poolSize) < t.maxPoolSize {
			conn, err := t.NewConnection()
			if err != nil {
				return t.GetConnection()
			}
			return conn
		}
		return <-t.connPool
	}
}

func (t *TorDns) ReleaseConnection(conn net.Conn) {
	t.connPool <- conn
}

func (t *TorDns) NewConnection() (net.Conn, error) {
	conn, err := t.Proxy.Dial("tcp", t.hiddenService)
	if err != nil {
		return nil, err
	}
	atomic.AddInt32(&t.poolSize, 1)
	return conn, nil
}

func (t TorDnsPlugin) Name() string {
	return "tordns"
}

func (t TorDnsPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	for i := 0; i < t.maxRetries; i++ {
		conn := t.GetConnection()
		ret, err := t.serve(ctx, conn, w, r)
		if err != nil {
			atomic.AddInt32(&t.poolSize, -1)
			conn.Close()
		} else {
			defer t.ReleaseConnection(conn)
			return ret, err
		}
	}

	return dns.RcodeServerFailure, errors.New("too many retries..")
}

func (t TorDnsPlugin) serve(ctx context.Context, conn net.Conn, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	co := &dns.Conn{Conn: conn}

	err := co.WriteMsg(r)
	if err != nil {
		log.Debugf("exchange error: %s\n", err)
		return dns.RcodeServerFailure, err
	}

	out, err := co.ReadMsg()
	if err != nil {
		log.Debugf("read error: %s\n", err)
		return dns.RcodeServerFailure, err
	}
	w.WriteMsg(out)
	return dns.RcodeSuccess, nil
}