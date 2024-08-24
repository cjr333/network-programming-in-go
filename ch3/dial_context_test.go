package ch3

import (
	"context"
	"io"
	"net"
	"syscall"
	"testing"
	"time"
)

func listen(t *testing.T) (net.Listener, error) {
	// 랜덤 포트에 리스너 생성
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	go func() {
		// 하나의 연결만을 수락합니다.
		conn, err := listener.Accept()
		if err != nil {
			t.Log(err)
			return
		}

		defer func() {
			conn.Close()
		}()
		t.Logf("connected on local: %q, remote: %q", conn.LocalAddr(), conn.RemoteAddr())

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					t.Error(err)
				}
				return
			}
			t.Logf("received: %q", buf[:n])
		}
	}()
	return listener, nil
}

func TestDialContext(t *testing.T) {
	listener, err := listen(t)
	if err != nil {
		t.Fatal(err)
	}
	dl := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer // DialContext는 Dialer의 메서드입니다.
	d.Control = func(_, addr string, _ syscall.RawConn) error {
		// 컨텍스트의 데드라인이 지나기 위해 충분히 긴 시간동안 대기합니다.
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
		//return &net.DNSError{
		//	Err:         "not time out error",
		//	Name:        addr,
		//	Server:      "127.0.0.1",
		//	IsTimeout:   false,
		//	IsTemporary: false,
		//}
	}
	//conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
	conn, err := d.DialContext(ctx, "tcp", listener.Addr().String())
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}
}
