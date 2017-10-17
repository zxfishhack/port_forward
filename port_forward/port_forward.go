package port_forward

import (
	"net"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type PortForward struct {
	addr string
	remote string
	isHttp bool
	listen net.Listener
	httpServer *http.Server
	proxy *httputil.ReverseProxy
	closeCh chan bool
}

func NewPortForward(listen string, remote string, isHttp bool) *PortForward {
	pf := &PortForward{
		addr: listen,
		remote: remote,
		isHttp: isHttp,
		closeCh: make(chan bool),
	}
		var err error
		pf.listen, err = net.Listen("tcp", listen)
		if err != nil {
			return nil
		}
		if pf.isHttp {
			pf.httpServer = &http.Server{
				Handler: pf,
			}
			remote, err := url.Parse("http://" + pf.remote)
			if err != nil {
				 panic(err)
			}
			lPort := "80"
			idx := strings.Index(pf.addr, ":")
			if idx != -1 {
				lPort = pf.addr[idx+1:]
			}
			pf.proxy = httputil.NewSingleHostReverseProxy(remote)
			pf.proxy.ModifyResponse = func(res *http.Response) error{
				location := res.Header.Get("Location")
				url, err := url.Parse(location)
				if location != "" && err == nil {
					idx := strings.Index(url.Host, ":")
					if idx == -1 {
						url.Host = url.Host + ":" + lPort;
					} else {
						url.Host = url.Host[:idx] + ":" + lPort;
					}
					res.Header.Set("Location", url.String())
				}
				return nil
			}
			log.Printf("http forward %v->%v", listen, remote)
			go pf.httpServer.Serve(pf.listen)
		} else {
			go pf.serve()
		}
	return pf
}

func (pf *PortForward)Close() {
	close(pf.closeCh)
	if (pf.isHttp) {
		pf.httpServer.Handler = pf
		pf.httpServer.Close()
	} else {
		pf.listen.Close()
	}
}

func (pf *PortForward) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v", r.URL)
	pf.proxy.ServeHTTP(w, r)
}

func (pf *PortForward)serve() {
	connCh := make(chan net.Conn)
	go func(){
		for {
			conn, err := pf.listen.Accept()
			if err != nil {
				break
			}
			connCh <- conn
		}
	}()
	select {
	case conn := <-connCh:
		go pf.forward(conn)
	case <-pf.closeCh:
	}
}

func (pf *PortForward)forward(conn net.Conn){
	rc, err := net.Dial("tcp", pf.remote)
	if err != nil {
		conn.Close()
		return
	}
	errCh := make(chan bool, 2)
	log.Printf("forwarding %v->%v, client: %v", pf.addr, pf.remote, conn.RemoteAddr())
	forward := func(src, dst net.Conn) {
		buf := make([]byte, 256)
		for {
			n, err := src.Read(buf)
			if err != nil || n == 0 {
				errCh <- true
				break
			}
			dst.Write(buf[:n])
		}
	}
	go forward(rc, conn)
	go forward(conn, rc)
	select {
	case <-pf.closeCh:
	case <-errCh:
	}
	rc.Close()
	conn.Close()
}
