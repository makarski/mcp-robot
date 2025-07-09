package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/makarski/mcp-robot/io"
	"github.com/makarski/mcp-robot/spec"
)

type StdioServer struct {
	*server
}

func (s *StdioServer) ListenAndServe() error {
	// fmt.Println("stared stdio server")

	w := os.Stdout
	decoder := json.NewDecoder(os.Stdin)
	for {
		var rpcReq spec.Request[int]
		if err := decoder.Decode(&rpcReq); err != nil {
			if err.Error() == "EOF" {
				// fmt.Println("EOF reached, exiting")
				return nil
			}

			rw := io.NewResponseWriter(w, 0)
			rw.WriteError(spec.ErrorCodeParseError, fmt.Sprintf("Failed to decode request: %s", err))
			continue
		}

		rw := io.NewResponseWriter(w, rpcReq.ID)
		handler, err := s.resolveHandler(rpcReq)
		if err != nil {
			errCode := spec.ErrorCodeInternalError
			pe, ok := err.(*spec.ProtocolError)
			if ok {
				errCode = pe.Code
			}

			rw.WriteError(errCode, fmt.Sprintf("Failed to resolve handler: %s", err))
			continue
		}

		handler.ServeRPC(w, &rpcReq)
	}
}
