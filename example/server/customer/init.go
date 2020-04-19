package customer

import (
	"net/http"
	"net/http/pprof"
)

func init() {
	go func() {
		addr := "localhost:9022"
		pprofServeMux := http.NewServeMux()
		pprofServeMux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		http.ListenAndServe(addr, pprofServeMux)
	}()
}
