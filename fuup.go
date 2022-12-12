package fuup

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PIngBZ/kcp-go/v5"
	"github.com/PIngBZ/socks5"
	"github.com/xtaci/smux"
)

type Fuup struct {
	asServer    bool
	allowProxy  bool
	fakeSubNet  string
	localSubNet string

	smux   atomic.Value
	kcp    atomic.Value
	socks5 atomic.Value
	die    atomic.Value
}

func NewFuup(asServer, allowProxy bool, listenSocks, fakeSubNet, localSubNet string) *Fuup {
	h := &Fuup{}
	h.asServer = asServer
	h.allowProxy = allowProxy
	h.fakeSubNet = fakeSubNet
	h.localSubNet = localSubNet

	if len(listenSocks) > 0 {
		go h.localListen(listenSocks)
	}

	return h
}

func (h *Fuup) HandleKCP(conn *kcp.UDPSession) {
	h.Close()

	log.Println("Fuup HandleKCP")
	defer log.Println("Fuup HandleKCP exit")

	die := make(chan struct{})
	h.die.Store(&die)

	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(1, 5, 10, 1)
	conn.SetWindowSize(10240, 10240)
	conn.SetMtu(470)
	conn.SetACKNoDelay(true)

	var s *smux.Session
	if h.asServer {
		s, _ = smux.Server(conn, SmuxConfig())
	} else {
		s, _ = smux.Client(conn, SmuxConfig())
	}

	h.kcp.Store(conn)
	h.smux.Store(s)

	if h.allowProxy {
		h.serveProxy(&SmuxServerWrapper{Session: s})
	} else {
		log.Println("Fuuup Not allow proxy")
		<-die
	}
}

func (h *Fuup) Close() {
	log.Println("Fuup Close")

	if old := h.socks5.Swap((*socks5.Server)(nil)); old != nil && old.(*socks5.Server) != nil {
		old.(*socks5.Server).Close()
	}

	if old := h.smux.Swap((*smux.Session)(nil)); old != nil && old.(*smux.Session) != nil {
		old.(*smux.Session).Close()
	}

	if old := h.kcp.Swap((*kcp.UDPSession)(nil)); old != nil && old.(*kcp.UDPSession) != nil {
		old.(*kcp.UDPSession).Close()
	}

	if old := h.die.Swap((*chan struct{})(nil)); old != nil && old.(*chan struct{}) != nil {
		close(*old.(*chan struct{}))
	}
}

func (h *Fuup) serveProxy(listener net.Listener) {
	log.Println("Fuup serveProxy")
	defer log.Println("Fuup serveProxy exit")

	socks5 := &socks5.Server{
		Addr:                   "1.1.1.1:1",
		Authenticators:         nil,
		DisableSocks4:          true,
		Transporter:            nil,
		DialTimeout:            time.Second * 5,
		HandshakeReadTimeout:   time.Second * 5,
		HandshakeWriteTimeout:  time.Second * 5,
		CallbackAfterHandshake: h.callbackAfterHandshake,
	}
	h.socks5.Store(socks5)

	socks5.Serve(listener)
}

func (h *Fuup) localListen(addr string) {
	log.Println("Fuup localListen " + addr)
	defer log.Println("Fuup localListen exit " + addr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	CheckError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckError(err)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Fuup local socks5 AcceptTCP exit: %+v\n", err)
			return
		}

		log.Printf("Fuup local socks5 AcceptTCP: %s\n", conn.RemoteAddr().String())

		if s := h.smux.Load(); s != nil {
			stream, err := s.(*smux.Session).OpenStream()
			if err != nil {
				conn.Close()
				log.Printf("Fuup local socks5 tunnel OpenStream error: %+v\n", err)
			} else {
				go Transfer(conn, stream)
			}
		} else {
			conn.Close()
			log.Println("Fuup local socks5 no tunnel connection")
		}
	}
}

func (h *Fuup) callbackAfterHandshake(srv *socks5.Server, req *socks5.Request) bool {
	switch req.Address.ATYPE {
	case socks5.IPV4_ADDRESS:
		addr := req.Address.Addr
		dst := fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
		if strings.HasPrefix(dst, h.fakeSubNet) {
			dst = strings.Replace(dst, h.fakeSubNet, h.localSubNet, 1)
			req.Address.Addr = net.ParseIP(dst)
		}
	case socks5.DOMAINNAME:
		dst := string(req.Address.Addr)
		if strings.HasPrefix(dst, h.fakeSubNet) {
			dst = strings.Replace(dst, h.fakeSubNet, h.localSubNet, 1)
			req.Address.Addr = []byte(dst)
		}
	}

	return true
}
