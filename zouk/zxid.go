package zouk

func (z *Zxid) LessThan(other *Zxid) bool {
	return z.Epoch < other.Epoch || (z.Epoch == other.Epoch && z.Counter < other.Counter)
}

func (z *Zxid) Equal(other *Zxid) bool {
	return z.Epoch == other.Epoch && z.Counter == other.Counter
}

func (z *Zxid) GreaterThan(other *Zxid) bool {
	return z.Epoch > other.Epoch || (z.Epoch == other.Epoch && z.Counter > other.Counter)
}
