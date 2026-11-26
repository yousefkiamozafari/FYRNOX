package vpnclient

import (
	"net"
	"sync"
	"time"
)

type fragmentingConn struct {
	net.Conn
	mu             sync.Mutex
	firstWriteDone bool
	mode           string
}

// NewFragmentingConn wraps c to fragment the initial TLS ClientHello packet across TCP frames,
// helping bypass DPI SNI filters.
func NewFragmentingConn(c net.Conn, mode string) net.Conn {
	return newFragmentingConn(c, mode)
}

func newFragmentingConn(c net.Conn, mode string) *fragmentingConn {
	if mode == "" {
		mode = "sni"
	}
	return &fragmentingConn{
		Conn: c,
		mode: mode,
	}
}

func (fc *fragmentingConn) Write(b []byte) (int, error) {
	fc.mu.Lock()
	if fc.firstWriteDone {
		fc.mu.Unlock()
		return fc.Conn.Write(b)
	}
	fc.firstWriteDone = true
	return fc.writeInitial(b)
}

func (fc *fragmentingConn) writeInitial(b []byte) (int, error) {
	defer fc.mu.Unlock()

	// Check if this write looks like a TLS ClientHello
	// TLS record header: 0x16 (Handshake), Version (0x03, 0x01..0x03)
	// Handshake type at offset 5: 0x01 (ClientHello)
	if len(b) > 9 && b[0] == 0x16 && b[1] == 0x03 && b[5] == 0x01 {
		switch fc.mode {
		case "first_byte":
			n1, err := fc.Conn.Write(b[:1])
			if err != nil {
				return n1, err
			}
			time.Sleep(8 * time.Millisecond)
			n2, err := fc.Conn.Write(b[1:])
			return n1 + n2, err

		case "sni", "chunks":
			fallthrough
		default:
			split1 := 5
			split2 := 35
			if len(b) > split2 {
				n1, err := fc.Conn.Write(b[:split1])
				if err != nil {
					return n1, err
				}
				time.Sleep(6 * time.Millisecond)

				n2, err := fc.Conn.Write(b[split1:split2])
				if err != nil {
					return n1 + n2, err
				}
				time.Sleep(6 * time.Millisecond)

				n3, err := fc.Conn.Write(b[split2:])
				return n1 + n2 + n3, err
			} else if len(b) > split1 {
				n1, err := fc.Conn.Write(b[:split1])
				if err != nil {
					return n1, err
				}
				time.Sleep(6 * time.Millisecond)

				n2, err := fc.Conn.Write(b[split1:])
				return n1 + n2, err
			}
		}
	}

	return fc.Conn.Write(b)
}
