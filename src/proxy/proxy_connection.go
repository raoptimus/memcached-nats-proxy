package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	nats "github.com/nats-io/go-nats-streaming"
)

var (
	crlf    = []byte("\r\n")
	noreply = []byte("noreply")
)

var (
	StatusEnd       = []byte("END\r\n")
	StatusStored    = []byte("STORED\r\n")
	StatusNotStored = []byte("NOT_STORED\r\n")
	StatusVersion   = []byte("VERSION 1.0.0\r\n")
	StatPattern     = "STAT %s %v\r\n"
)

type connect struct {
	net      net.Conn
	buffer   *bufio.ReadWriter
	natsConn nats.Conn
}

func (conn *connect) serve() {
	address := conn.net.RemoteAddr().String()
	{
		connectionsInc(address)
	}
	for {
		switch err := conn.handle(); {
		case err != nil:
			connectionsDec(address)
			{
				conn.net.Close()
			}
			return
		default:
			if err := conn.buffer.Flush(); err != nil {
				return
			}
		}
	}
}

func (conn *connect) handle() error {
	line, _, err := conn.buffer.ReadLine()
	if err != nil || len(line) < 4 {
		return io.EOF
	}
	switch line[0] {
	case 'g': // get
		keys := strings.Fields(string(line[4:]))
		if len(keys) == 0 {
			return io.EOF
		}
		for _, key := range keys {
			conn.buffer.Write([]byte("VALUE " + key + " 0 13\r\n"))
			conn.buffer.Write([]byte("not supported"))
			conn.buffer.Write(crlf)
		}
		conn.buffer.Write(StatusEnd)
	case 'q': // quit
		return io.EOF
	case 's':
		switch line[1] {
		case 'e': // set
			var (
				fields    = bytes.Fields(line[4:])
				subject   = string(fields[0])
				length, _ = strconv.Atoi(string(fields[3]))
				value     = make([]byte, length+2)
				n, err    = conn.buffer.Read(value)
			)
			if err != nil && err != io.EOF {
				return err
			}
			if !bytes.HasSuffix(value, crlf) {
				return err
			}
			switch err := conn.natsConn.Publish(subject, value[:n]); {
			case err != nil:
				conn.net.Write(StatusNotStored)
			default:
				conn.net.Write(StatusStored)
				reqProcessedInc()
			}
		case 't': // stats
			fmt.Fprintf(conn.buffer, StatPattern, "num_goroutine", runtime.NumGoroutine())
			fmt.Fprintf(conn.buffer, StatPattern, "cmd_set", atomic.LoadInt64(&reqProcessed))
			fmt.Fprintf(conn.buffer, StatPattern, "curr_connections", atomic.LoadInt64(&currentConnections))
			fmt.Fprintf(conn.buffer, StatPattern, "total_connections", atomic.LoadInt64(&totalConnections))
			conn.buffer.Write(StatusEnd)
		}
	case 'v': // version
		conn.net.Write(StatusVersion)
	default:
		return io.EOF
	}
	return nil
}