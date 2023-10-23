package zouk

type ZxidFragment struct {
	Epoch   int
	Counter int
}

func (z *Zxid) Data() ZxidFragment {
	return ZxidFragment{int(z.Epoch), int(z.Counter)}
}

func (z ZxidFragment) LessThan(other ZxidFragment) bool {
	return z.Epoch < other.Epoch || (z.Epoch == other.Epoch && z.Counter < other.Counter)
}

func (z ZxidFragment) Equal(other ZxidFragment) bool {
	return z.Epoch == other.Epoch && z.Counter == other.Counter
}

func (z ZxidFragment) GreaterThan(other ZxidFragment) bool {
	return z.Epoch > other.Epoch || (z.Epoch == other.Epoch && z.Counter > other.Counter)
}

func (z ZxidFragment) Raw() *Zxid {
	return &Zxid{Epoch: int64(z.Epoch), Counter: int64(z.Counter)}
}
