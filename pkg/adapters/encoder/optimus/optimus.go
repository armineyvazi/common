package optimus

import (
	"github.com/armineyvazi/common.git/pkg/ports"
	opPrime "github.com/pjebs/optimus-go"
)

type optimus struct {
	encoder opPrime.Optimus
}

func NewOptimusEncoder(conf Config) ports.Encoder {
	if conf.Prime == 0 {
		conf.Prime = 1580030173
	}
	if conf.ModInverse == 0 {
		conf.ModInverse = 59260789
	}
	if conf.Random == 0 {
		conf.Random = 1163945558
	}
	optimusPrimeEncoder := opPrime.New(conf.Prime, conf.ModInverse, conf.Random)
	return &optimus{
		encoder: optimusPrimeEncoder,
	}
}

func (o *optimus) Decode(input uint64) uint64 {
	return o.encoder.Decode(input)
}

func (o *optimus) Encode(input uint64) uint64 {
	return o.encoder.Encode(input)
}
