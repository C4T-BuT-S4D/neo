package cli

import (
	"bytes"

	"github.com/samber/lo"

	"github.com/c4t-but-s4d/neo/v2/proto/go/exploits"
)

func isBinary(data []byte) bool {
	return bytes.Equal(data[:4], []byte("\x7fELF"))
}

func getExploitFromState(state *exploits.ServerState, exploitID string) *exploits.ExploitState {
	exp, ok := lo.Find(state.Exploits, func(s *exploits.ExploitState) bool {
		return s.ExploitId == exploitID
	})
	if !ok {
		return nil
	}
	return exp
}
