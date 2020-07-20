package proxy

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Host            string
	Protocol        string
	CertFile        string
	KeyFile         string
	CredentialsFile string
}

func Start(config Config) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	server := &http.Server{
		Addr: config.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL)
			if r.Method == http.MethodConnect {
				handleConnect(w, r)
			} else {
				handleDirect(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	log.Printf("Started %s proxy on %s", config.Protocol, config.Host)

	if config.Protocol == "https" {
		log.Fatal(server.ListenAndServeTLS(config.CertFile, config.KeyFile))
	} else {
		log.Fatal(server.ListenAndServe())
	}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)

	if !ok {
		log.Println("http hijacking is not supported")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	conn, _, err := hijacker.Hijack()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go transfer(conn, targetConn)
	go transfer(targetConn, conn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleDirect(w http.ResponseWriter, r *http.Request) {
	response, err := http.DefaultTransport.RoundTrip(r)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	defer response.Body.Close()

	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}
