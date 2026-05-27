package theme

import (
	"os"
	"strings"
)

type Capability int

const (
	CapTruecolor Capability = iota
	Cap256
	Cap16
)

var ColorCapability = CapTruecolor

func DetectCapability() Capability {
	ct := os.Getenv("COLORTERM")
	if strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit") {
		return CapTruecolor
	}
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") || strings.Contains(term, "256") {
		return Cap256
	}
	return Cap16
}

func init() {
	ColorCapability = DetectCapability()
}
