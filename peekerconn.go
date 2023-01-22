// This package allows you to peek into a net.Conn "before reading".
// It is useful when you have to inspect a packet and route it
// depending on the content. The magic is that it reassembles
// the packet before reading from it.
// It is not thread safe until you Read at least as much as you Peek-ed.
package peekerconn

import (
	//"fmt"
	"bytes"
	"io"
	"net"
)

// make this its standalone package tomorrow
type peekerConn struct {
	net.Conn
	multiread io.Reader
	lenbuf    int
}

func NewPeekerConn(conn net.Conn) *peekerConn {
	return &peekerConn{Conn: conn}
}

// Implements a Peek function to inspect traffic on the fly
// You should have pconn.Read > the sum of peeked bytes before
// you hand it off into unknown space. You can also reuse the
// non peekerConn wrapped net.Conn after that, but you will
// loose data up to max peeked.
func (pconn *peekerConn) Peek(p []byte) (int, error) {
	buf := make([]byte, len(p))
	var numread int
	var err error
	if pconn.multiread != nil {
		//fmt.Println("peekerConn: re-read old stuff")
		numread, err = io.ReadFull(pconn.multiread, buf)
		pconn.multiread = io.MultiReader(bytes.NewReader(buf), pconn.multiread)
		pconn.lenbuf += numread
	} else {
		//fmt.Println("peekerConn: unpeeked data")
		numread, err = pconn.Conn.Read(buf)
		pconn.multiread = io.MultiReader(bytes.NewReader(buf), pconn.Conn)
		pconn.lenbuf = numread
	}
	copy(p, buf)
	//fmt.Println("peekerConn: len:", numread)
	return numread, err
}

// Implements the default Read interface, but it is not
// thread safe when pconn.multiread != nil.
func (pconn *peekerConn) Read(p []byte) (int, error) {
	if pconn.multiread != nil {
		//fmt.Println("peekerConn: hasmultiread")
		numread, err := pconn.multiread.Read(p)
		pconn.lenbuf -= numread //reduce our restbuffer
		if pconn.lenbuf <= 0 {
			//we have read all of it
			//fmt.Println("peekerConn: multiread empty")
			pconn.multiread = nil
			pconn.lenbuf = 0
		}
		//fmt.Println("peekerConn: multiread bytes left:", pconn.lenbuf)
		return numread, err
	}
	return pconn.Conn.Read(p)
}
