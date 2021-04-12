package types

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestPeerID_MarshalJSON(t *testing.T) {
	peerId1 := NewRandPeerID()
	json1, err := json.Marshal(peerId1)
	require.NoError(t, err)

	log.Println("peerId1", peerId1)
	log.Println("json1", string(json1))

	peerId2 := NewRandPeerID()
	require.NotEqual(t, peerId1, peerId2)

	err = json.Unmarshal(json1, peerId2)
	require.NoError(t, err)
	require.Equal(t, peerId1, peerId2)

	log.Println("peerId2", peerId1)
}
