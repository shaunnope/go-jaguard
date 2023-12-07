package zouk

import (
	"errors"
	"fmt"

	"github.com/shaunnope/go-jaguard/utils"
)

// dataclass for zxid
type ZxidFragment struct {
	Epoch   int
	Counter int
}

func (z *Zxid) Extract() ZxidFragment {
	return ZxidFragment{Epoch: int(z.Epoch), Counter: int(z.Counter)}
}

func (z ZxidFragment) String() string {
	return fmt.Sprintf("{%d %d}", z.Epoch, z.Counter)
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

// Increment epoch and reset counter
func (z ZxidFragment) Next() ZxidFragment {
	return ZxidFragment{Epoch: z.Epoch + 1, Counter: 0}
}

// Increment counter
func (z ZxidFragment) Inc() ZxidFragment {
	return ZxidFragment{Epoch: z.Epoch, Counter: z.Counter + 1}
}

// Convert to raw Zxid
func (z ZxidFragment) Raw() *Zxid {
	return &Zxid{Epoch: int64(z.Epoch), Counter: int64(z.Counter)}
}

func (z *ZxidFragment) Unmarshal(data []byte) error {
	if len(data) != 16 {
		return errors.New("invalid zxid length")
	}
	z.Epoch = utils.UnmarshalInt(data[0:8])
	z.Counter = utils.UnmarshalInt(data[8:16])
	return nil
}

func (z *ZxidFragment) Marshal() []byte {
	data := make([]byte, 16)
	copy(data[0:8], utils.MarshalInt(z.Epoch))
	copy(data[8:16], utils.MarshalInt(z.Counter))
	return data
}

type TransactionFragment struct {
	Zxid      ZxidFragment
	Path      string
	Data      []byte
	Flags     *Flag
	Type      OperationType
	Committed bool
}

func (t TransactionFragment) String() string {
	return fmt.Sprintf("{%d %d} %s @ %s %v", t.Zxid.Epoch, t.Zxid.Counter, t.Type, t.Path, t.Data)
}

type TransactionFragments struct {
	Transactions []TransactionFragment
	LastCommitId int
}

func (ts TransactionFragments) String() string {
	return fmt.Sprintf("%d %v", ts.LastCommitId, ts.Transactions)
}

func (ts TransactionFragments) Len() int {
	return len(ts.Transactions)
}

func (ts TransactionFragments) Raw() []*Transaction {
	res := make([]*Transaction, len(ts.Transactions))
	for i, t := range ts.Transactions {
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

func (ts *TransactionFragments) LastCommitZxid() ZxidFragment {
	if ts.LastCommitId == -1 || ts.LastCommitId >= len(ts.Transactions) {
		return ZxidFragment{}
	}
	return ts.Transactions[ts.LastCommitId].Zxid
}

func (ts *TransactionFragments) Set(hist []TransactionFragment) {
	// TODO: update non-volatile memory
	ts.Transactions = hist
	ts.LastCommitId = -1
}

// TODO: consider if this is needed
func (ts *TransactionFragments) CommitAll() {
	for i := 0; i < len(ts.Transactions); i++ {
		ts.Transactions[i].Committed = true
	}
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

func (t *Transaction) ExtractLog() TransactionFragment {
	return TransactionFragment{
		Zxid: t.Zxid.Extract(),
		Path: t.Path,
		Type: t.Type,
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
