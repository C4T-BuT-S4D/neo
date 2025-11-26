package cli

import (
	"bytes"

	"github.com/samber/lo"

	"github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
)

func isBinary(data []byte) bool {
	return bytes.Equal(data[:4], []byte("\x7fELF"))
}

func getExploitFromState(state *exploits.ServerState, exploitID string) *exploits.ExploitState {
	return lo.FindOrElse(state.GetExploits(), nil, func(s *exploits.ExploitState) bool {
		return s.GetExploitId() == exploitID
	})
}
