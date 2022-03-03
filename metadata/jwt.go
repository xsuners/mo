package metadata

import (
	"github.com/golang-jwt/jwt"
)

var _ jwt.Claims = (*Metadata)(nil)

// TODO
func (md *Metadata) Valid() error {
	return nil
}
