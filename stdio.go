package mcprobot

import (
	"encoding/json"
	"fmt"
	"os"
)

type StdioServer struct {
	*server
}

func (s *StdioServer) ListenAndServe() error {
	// fmt.Println("stared stdio server")

	w := os.Stdout
	decoder := json.NewDecoder(os.Stdin)
	for {
		var rpcReq Request[int]
		if err := decoder.Decode(&rpcReq); err != nil {
			if err.Error() == "EOF" {
				// fmt.Println("EOF reached, exiting")
				return nil
			}

			rw := NewResponseWriter(w, 0)
			rw.WriteError(ErrorCodeParseError, fmt.Sprintf("Failed to decode request: %s", err))
			continue
		}

		rw := NewResponseWriter(w, rpcReq.ID)
		handler, err := s.resolveHandler(rpcReq)
		if err != nil {
			errCode := ErrorCodeInternalError
			pe, ok := err.(*ProtocolError)
			if ok {
				errCode = pe.Code
			}

			rw.WriteError(errCode, fmt.Sprintf("Failed to resolve handler: %s", err))
			continue
		}

		handler.ServeRPC(w, &rpcReq)
	}
}
