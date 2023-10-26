package zouk

type ZxidFragment struct {
	Epoch   int
	Counter int
}

func (z *Zxid) Extract() ZxidFragment {
	return ZxidFragment{Epoch: int(z.Epoch), Counter: int(z.Counter)}
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

type TransactionFragment struct {
	Zxid  ZxidFragment
	Path  string
	Data  []byte
	Flags string
	Type  int
}

func (t *Transaction) Extract() TransactionFragment {
	return TransactionFragment{
		Zxid:  t.Zxid.Extract(),
		Path:  t.Path,
		Data:  t.Data,
		Flags: t.Flags,
		Type:  int(t.Type),
	}
}

type Transactions []*Transaction

func (ts Transactions) ExtractAll() []TransactionFragment {
	res := make([]TransactionFragment, len(ts))
	for i, t := range ts {
		res[i] = t.Extract()
	}
	return res
}
