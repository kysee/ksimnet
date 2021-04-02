package knetsim_test

import "testing"
import "github.com/knetsim"

func TestAddHost(t *testing.T) {
	knetsim.NewServer("1.1.1.1", 10)
	knetsim.NewServer("1.1.1.2", 10)

	knetsim.PrintServers()
}
