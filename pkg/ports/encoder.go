package ports

type Encoder interface {
	Encode(input uint64) uint64
	Decode(input uint64) uint64
}
