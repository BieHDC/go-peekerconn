// This package allows you to peek into a net.Conn "before reading".
// It is useful when you have to inspect a packet and route it
// depending on the content.
package peekerconn

import (
	"bufio"
	"net"
)

type peekerConn struct {
	r        *bufio.Reader
	net.Conn
}

func NewPeekerConn(c net.Conn) *peekerConn {
	return &peekerConn{bufio.NewReader(c), c}
}

func (b *peekerConn) Peek(p []byte) (int, error) {
	bytes, err := b.r.Peek(len(p))
	copy(p, bytes)
	return len(bytes), err
}

func (b *peekerConn) Read(p []byte) (int, error) {
	return b.r.Read(p)
}
