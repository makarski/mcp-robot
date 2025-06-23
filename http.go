package mcprobot

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPServer struct {
	*server
}

func (s *HTTPServer) ListenAndServe(pattern string) func(addr string, mux *http.ServeMux) error {
	fmt.Println("entered listen and serve")

	return func(addr string, mux *http.ServeMux) error {
		if mux == nil {
			fmt.Println("mux is nil, creating a new ServeMux")
			http.HandleFunc("POST /"+pattern, s.entrypoint)
			http.HandleFunc("OPTIONS /"+pattern, optionsHandler)
			return http.ListenAndServe(addr, nil)
		} else {
			fmt.Println("mux is not nil, using provided ServeMux")
			mux.HandleFunc("POST /"+pattern, s.entrypoint)
			mux.HandleFunc("OPTIONS /"+pattern, optionsHandler)
			return http.ListenAndServe(addr, mux)
		}
	}
}

func (s *HTTPServer) entrypoint(w http.ResponseWriter, req *http.Request) {
	fmt.Println("entered entrypoint")

	var rpcReq Request[int]
	if err := json.NewDecoder(req.Body).Decode(&rpcReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	handler, err := s.resolveHandler(rpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve handler: %s", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	handler.ServeRPC(w, &rpcReq)
}

func optionsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
	w.WriteHeader(http.StatusOK)
}
