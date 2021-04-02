package knetsim_test

import "testing"
import "github.com/knetsim"

func TestAddHost(t *testing.T) {
	knetsim.NewHost("1.1.1.1", 10)
	knetsim.NewHost("1.1.1.2", 10)

	knetsim.PrintHosts()
}
