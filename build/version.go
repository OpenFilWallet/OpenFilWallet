package build

import "fmt"

var CurrentCommit string

var version = "v1.0.0-rc1"

func Version() string {
	return fmt.Sprintf("%s%s", version, CurrentCommit)
}
