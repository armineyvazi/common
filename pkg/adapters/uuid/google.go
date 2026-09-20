package uuid

import (
	"github.com/armineyvazi/common.git/pkg/ports"
	"github.com/google/uuid"
)

type google struct{}

func New() ports.UUID {
	return &google{}
}

func (g *google) GenV4() (string, error) {
	_uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return _uuid.String(), nil
}
