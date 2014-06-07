package simple

import (
	"net/http"
	"runtime"
	"time"
)

type Server struct {
	*RestRouter
	Address   string
	Timeout   time.Duration
	StaticDir string
}

/* sever ---------------- */

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func NewServer(addr string) *Server {
	return &Server{
		Address:    addr,
		Timeout:    time.Second * 3,
		RestRouter: NewRestRouter("/api/"),
	}
}

func (this *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.Handle(this.Pattern, this)

	if this.StaticDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(this.StaticDir)))
	}

	server := &http.Server{
		Addr:        this.Address,
		ReadTimeout: this.Timeout,
		Handler:     mux,
	}

	return server.ListenAndServe()
}
