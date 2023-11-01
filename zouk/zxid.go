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

func (z ZxidFragment) Inc() ZxidFragment {
	return ZxidFragment{Epoch: z.Epoch, Counter: z.Counter + 1}
}

func (z ZxidFragment) Raw() *Zxid {
	return &Zxid{Epoch: int64(z.Epoch), Counter: int64(z.Counter)}
}

const (
	Ephemeral = 1 << iota
	Sequential
	Regular
)

type TransactionFragment struct {
	Zxid  ZxidFragment
	Path  string
	Data  []byte
	Flags uint64
	Type  OperationType
}

type TransactionFragments []TransactionFragment

func (ts TransactionFragments) Raw() []*Transaction {
	res := make([]*Transaction, len(ts))
	for i, t := range ts {
		res[i] = &Transaction{
			Zxid:  t.Zxid.Raw(),
			Path:  t.Path,
			Data:  t.Data,
			Flags: t.Flags,
			Type:  t.Type,
		}
	}
	return res
}

func (t *Transaction) Extract() TransactionFragment {
	return TransactionFragment{
		Zxid:  t.Zxid.Extract(),
		Path:  t.Path,
		Data:  t.Data,
		Flags: t.Flags,
		Type:  t.Type,
	}
}

func (t *Transaction) WithZxid(z ZxidFragment) *Transaction {
	return &Transaction{
		Zxid:  z.Raw(),
		Path:  t.Path,
		Data:  t.Data,
		Flags: t.Flags,
		Type:  t.Type,
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
