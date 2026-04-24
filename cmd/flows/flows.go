package flows

import (
	"github.com/lucaskatayama/oauth2-cli/internal/flow"
)

func init() {
	flow.Register(&AuthCodeFlow{})
	flow.Register(&ClientCredentialsFlow{})
}
