package peekerconn

import (
	"bufio"
	"bytes"
	"net"
	"testing"
	//"io"
	"time"
)

var endtransmission = []byte("please close the connection\n")

func disconnectEchoServer(t *testing.T, conn net.Conn) {
	_, err := conn.Write(endtransmission)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}
	conn.Close()
}

func connectEchoServer(t *testing.T, echoserver net.Addr) (net.Conn, func(), error) {
	conn, err := net.Dial("tcp", echoserver.String())
	if err != nil {
		return nil, nil, err
	}

	disconnect := func() {
		disconnectEchoServer(t, conn)
	}

	return conn, disconnect, nil
}

// Implements a simple helper echo server for testing.
// Echos back when encountering a \n, so you must provide it.
func spawnEchoServer(t *testing.T, addr string) (net.Addr, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	//t.Logf("Helper Echo Server listening on: %s ", listener.Addr())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Fatal("failed to accept connection:", err)
				continue
			}
			//t.Log("incoming connection from", conn.RemoteAddr())

			go func(conn net.Conn) {
				for {
					reader := bufio.NewReader(conn)
					content, err := reader.ReadBytes(byte('\n'))
					if err != nil {
						t.Fatal("failed to read all data:", err)
						return
					}
					//t.Log("read data:", content)

					if bytes.Equal(content, endtransmission) {
						//t.Log("connection termination requested")
						conn.Close()
						return
					}

					_, err = conn.Write(content)
					if err != nil {
						t.Fatal("failed to write to conn:", err)
						return
					}
					//t.Log("replied to connection:", conn.RemoteAddr())
				}
			}(conn)
		}
	}()

	return listener.Addr(), nil
}

// Sanity Check Test
func testNormalOperation(t *testing.T, echoserver net.Addr) {
	conn, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	conn.SetDeadline(time.Now().Add(time.Second * 2))

	// Test normal operation
	normal := []byte("Using it as it is supposed to be used\n")
	_, err = conn.Write(normal)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(normal, content) {
		t.Fatal("expected", normal, "got", content)
	}
}

func testPeekOneTime(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("hello, this should be peeked once\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek := make([]byte, 5)
	_, err = conn.Peek(peek)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek, msgwithonepeek[:len(peek)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek)], "got", peek)
	} else {
		t.Log("peek once success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func testPeekTwiceEqualSize(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("AABBCCDDEEFFGGHHIIJJKKLLMMOOPPQQRRSSTTUUVVWWXXYYZZ\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek1 := make([]byte, 5)
	_, err = conn.Peek(peek1)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek1, msgwithonepeek[:len(peek1)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek1)], "got", peek1)
	} else {
		t.Log("peek1 success")
	}

	peek2 := make([]byte, 5)
	_, err = conn.Peek(peek2)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek2, msgwithonepeek[:len(peek2)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek2)], "got", peek2)
	} else {
		t.Log("peek2 success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func testPeekTwiceSmallerBigger(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("AABBCCDDEEFFGGHHIIJJKKLLMMOOPPQQRRSSTTUUVVWWXXYYZZ\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek1 := make([]byte, 5)
	numread, err := conn.Peek(peek1)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if numread != len(peek1) {
		t.Log("read only", numread, "when we wanted", len(peek1))
	}
	if !bytes.Equal(peek1, msgwithonepeek[:len(peek1)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek1)], "got", peek1)
	} else {
		t.Log("peek1 success")
	}

	peek2 := make([]byte, 10)
	numread, err = conn.Peek(peek2)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if numread != len(peek2) {
		t.Log("read only", numread, "when we wanted", len(peek2))
	}
	if !bytes.Equal(peek2, msgwithonepeek[:len(peek2)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek2)], "got", peek2)
	} else {
		t.Log("peek2 success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func testPeekTriceSmallerBiggerBigger(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("AABBCCDDEEFFGGHHIIJJKKLLMMOOPPQQRRSSTTUUVVWWXXYYZZ\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek1 := make([]byte, 5)
	numread, err := conn.Peek(peek1)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if numread != len(peek1) {
		t.Log("read only", numread, "when we wanted", len(peek1))
	}
	if !bytes.Equal(peek1, msgwithonepeek[:len(peek1)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek1)], "got", peek1)
	} else {
		t.Log("peek1 success")
	}

	peek2 := make([]byte, 10)
	numread, err = conn.Peek(peek2)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if numread != len(peek2) {
		t.Log("read only", numread, "when we wanted", len(peek2))
	}
	if !bytes.Equal(peek2, msgwithonepeek[:len(peek2)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek2)], "got", peek2)
	} else {
		t.Log("peek2 success")
	}

	peek3 := make([]byte, 15)
	numread, err = conn.Peek(peek3)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if numread != len(peek3) {
		t.Log("read only", numread, "when we wanted", len(peek3))
	}
	if !bytes.Equal(peek3, msgwithonepeek[:len(peek3)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek3)], "got", peek3)
	} else {
		t.Log("peek3 success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func testPeekTwiceBiggerSmaller(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("AABBCCDDEEFFGGHHIIJJKKLLMMOOPPQQRRSSTTUUVVWWXXYYZZ\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek1 := make([]byte, 10)
	_, err = conn.Peek(peek1)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek1, msgwithonepeek[:len(peek1)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek1)], "got", peek1)
	} else {
		t.Log("peek1 success")
	}

	peek2 := make([]byte, 5)
	_, err = conn.Peek(peek2)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek2, msgwithonepeek[:len(peek2)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek2)], "got", peek2)
	} else {
		t.Log("peek2 success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func testPeekTriceBiggerSmallerSmaller(t *testing.T, echoserver net.Addr) {
	connog, disconnect, err := connectEchoServer(t, echoserver)
	if err != nil {
		t.Fatal("failed to dial echo server:", err)
		return
	}
	defer disconnect()
	// Helps when we mess up
	connog.SetDeadline(time.Now().Add(time.Second * 2))

	//Upgrade our conn to a peekerConn
	conn := NewPeekerConn(connog)

	msgwithonepeek := []byte("AABBCCDDEEFFGGHHIIJJKKLLMMOOPPQQRRSSTTUUVVWWXXYYZZ\n")
	_, err = conn.Write(msgwithonepeek)
	if err != nil {
		t.Fatal("failed to write to conn:", err)
	}

	peek1 := make([]byte, 15)
	_, err = conn.Peek(peek1)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek1, msgwithonepeek[:len(peek1)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek1)], "got", peek1)
	} else {
		t.Log("peek1 success")
	}

	peek2 := make([]byte, 10)
	_, err = conn.Peek(peek2)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek2, msgwithonepeek[:len(peek2)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek2)], "got", peek2)
	} else {
		t.Log("peek2 success")
	}

	peek3 := make([]byte, 5)
	_, err = conn.Peek(peek3)
	if err != nil {
		t.Fatal("failed to peek:", err)
	}
	if !bytes.Equal(peek3, msgwithonepeek[:len(peek3)]) {
		t.Fatal("expected", msgwithonepeek[:len(peek3)], "got", peek3)
	} else {
		t.Log("peek3 success")
	}

	reader := bufio.NewReader(conn)
	content, err := reader.ReadBytes(byte('\n'))
	if err != nil {
		t.Fatal("failed to read from conn:", err)
	}

	if !bytes.Equal(msgwithonepeek, content) {
		t.Fatal("expected", msgwithonepeek, "got", content)
	}
}

func TestPeekerConn(t *testing.T) {
	echoserver, err := spawnEchoServer(t, "127.0.0.1:0")
	if err != nil {
		t.Fatal("failed to spawn echo server:", err)
	}

	t.Run("normal operation", func(t *testing.T) {
		testNormalOperation(t, echoserver)
	})
	t.Run("peek one time", func(t *testing.T) {
		testPeekOneTime(t, echoserver)
	})
	t.Run("peek twice of equal size", func(t *testing.T) {
		testPeekTwiceEqualSize(t, echoserver)
	})
	t.Run("peek twice smaller then bigger", func(t *testing.T) {
		testPeekTwiceSmallerBigger(t, echoserver)
	})
	t.Run("peek trice smaller then bigger then even bigger", func(t *testing.T) {
		testPeekTriceSmallerBiggerBigger(t, echoserver)
	})
	t.Run("peek twice bigger then smaller (not real world)", func(t *testing.T) {
		testPeekTwiceBiggerSmaller(t, echoserver)
	})
	t.Run("peek trice bigger then smaller then smaller (not real world)", func(t *testing.T) {
		testPeekTriceBiggerSmallerSmaller(t, echoserver)
	})
}
