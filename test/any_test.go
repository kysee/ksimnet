package any

import (
	"fmt"
	"net"
	"testing"
	"github.com/stretchr/testify/require"
)

func TestTCPAddr(t *testing.T) {
	ta, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1234")
	require.NoError(t, err)

	t.Log(ta.String())
	t.Log(ta.Zone)
	t.Log(ta.Network())
	t.Log(ta.IP)
	t.Log(ta.Port)
}

func TestSlice(t *testing.T) {
	s0 := make([]int, 5, 10)
	s1 := s0
	for i, _ := range s0 {
		s0[i] = 1
	}
	fmt.Println(len(s0), s0)
	fmt.Println(len(s1), s1)

	s0 = s0[:cap(s0)]
	for i, _ := range s0 {
		s0[i] = 2
	}
	fmt.Println(len(s0), s0)
	fmt.Println(len(s1), s1)

}
