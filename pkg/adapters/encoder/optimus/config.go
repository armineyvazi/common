package optimus

type Config struct {
	Prime      uint64 `mapstructure:"prime" config:"prime" default:"${OPTIMUS_PRIME}"`
	ModInverse uint64 `mapstructure:"mod_inverse" config:"mod_inverse" default:"${OPTIMUS_MOD_INVERSE}"`
	Random     uint64 `mapstructure:"random" config:"random" default:"${OPTIMUS_RANDOM}"`
}
