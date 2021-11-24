package proxy

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	ServerAddr       string
	MetricsAddr      string
	NatsPublishAsync bool
	NatsURL          string
	MaxInFlight      int
}

// New NATS Steaming proxy
func New(version string, options Options) (_ *Proxy, err error) {
	var (
		proxy = Proxy{
			version: version,
			address: options.ServerAddr,
			signals: make(chan os.Signal),
		}
		connOpts = []nats.Option{
			nats.MaxPingsOutstanding(10),
			nats.PingInterval(120 * time.Second),
			nats.RetryOnFailedConnect(true),
			nats.ReconnectWait(3 * time.Second),
			nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
				log.Warningf("Got disconnected! Reason:: %q\n", err)
			}),
			nats.ReconnectHandler(func(nc *nats.Conn) {
				log.Warningf("Got reconnected to %v!\n", nc.ConnectedUrl())
			}),
			nats.ClosedHandler(func(nc *nats.Conn) {
				log.Warningf("Connection closed. Reason: %q\n", nc.LastError())
			}),
		}
	)
	if proxy.nats.conn, err = nats.Connect(options.NatsURL, connOpts...); err != nil {
		return nil, err
	}
	if proxy.nats.js, err = proxy.nats.conn.JetStream(nats.PublishAsyncMaxPending(options.MaxInFlight)); err != nil {
		return nil, err
	}

	proxy.nats.async = options.NatsPublishAsync
	log.Infof("listen=%s, nats=%s, nats-publish-async=%t",
		options.ServerAddr,
		options.NatsURL,
		proxy.nats.async,
	)
	go metrics(options.MetricsAddr)
	go proxy.waitSignal()

	return &proxy, nil
}

type Proxy struct {
	version string
	address string
	signals chan os.Signal
	nats    struct {
		conn  *nats.Conn
		js    nats.JetStreamContext
		async bool
	}
}

func (p *Proxy) waitSignal() {
	signal.Notify(p.signals,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	select {
	case sig := <-p.signals:
		log.Infof("shutdown [%s]", sig)
		p.nats.conn.Close()
		{
			os.Exit(0)
		}
	}
}

func (p *Proxy) publish(subject string, data []byte) (err error) {
	switch {
	case p.nats.async:
		_, err = p.nats.js.PublishAsync(subject, data)
	default:
		_, err = p.nats.js.Publish(subject, data)
	}
	return err
}

// Listen announces on the local network address.
func (p *Proxy) Listen() error {
	listener, err := net.Listen("tcp", p.address)
	if err != nil {
		return err
	}
	for {
		if conn, err := listener.Accept(); err == nil {
			go (&connect{
				net:     conn,
				version: p.version,
				publish: p.publish,
				buffer: bufio.NewReadWriter(
					bufio.NewReaderSize(conn, 8*1024),
					bufio.NewWriter(conn),
				),
			}).serve()
		}
	}
}
