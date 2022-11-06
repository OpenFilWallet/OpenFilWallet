package build

import "fmt"

var CurrentCommit string

var version = "v0.0.1"

func Version() string {
	return fmt.Sprintf("%s%s", version, CurrentCommit)
}
