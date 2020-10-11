package target

import (
	"github.com/krilor/gossh/target/local"
	"github.com/krilor/gossh/target/rmt"
)

var _ Target = &rmt.Remote{}
var _ Target = &local.Local{}
