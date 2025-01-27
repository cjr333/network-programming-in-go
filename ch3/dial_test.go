package ch3

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 랜덤 포트에 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("bound to %q", listener.Addr())

	done := make(chan struct{})
	go func() {
		defer func() {
			done <- struct{}{}
		}()

		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()
				t.Logf("connected on local: %q, remote: %q", conn.LocalAddr(), conn.RemoteAddr())

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Write([]byte("abc"))
	conn.Close()
	<-done

	conn, err = net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Write([]byte("123"))
	conn.Close()
	<-done
	listener.Close()
	<-done
}
