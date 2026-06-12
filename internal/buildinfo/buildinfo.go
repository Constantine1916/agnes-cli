package buildinfo

import "runtime"

var Version = "0.0.4"

func UserAgent() string {
	return "agnes-cli/" + Version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
}
