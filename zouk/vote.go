package zouk

import (
	"errors"

	"github.com/shaunnope/go-jaguard/utils"
)

type VoteFragment struct {
	LastZxid ZxidFragment
	Id       int
}

func (v *VoteFragment) Unmarshal(data []byte) error {
	if len(data) != 24 {
		return errors.New("invalid vote length")
	}
	v.LastZxid = ZxidFragment{Epoch: utils.UnmarshalInt(data[0:8]), Counter: utils.UnmarshalInt(data[8:16])}
	v.Id = utils.UnmarshalInt(data[16:24])
	return nil
}

func (v *VoteFragment) Marshal() []byte {
	data := make([]byte, 24)
	copy(data[0:8], utils.MarshalInt(v.LastZxid.Epoch))
	copy(data[8:16], utils.MarshalInt(v.LastZxid.Counter))
	copy(data[16:24], utils.MarshalInt(v.Id))
	return data
}

func (v *Vote) Data() VoteFragment {
	return VoteFragment{v.LastZxid.Extract(), int(v.Id)}
}

func (v VoteFragment) LessThan(other VoteFragment) bool {
	return v.LastZxid.LessThan(other.LastZxid) || (v.LastZxid.Equal(other.LastZxid) && v.Id < other.Id)
}

func (v VoteFragment) Equal(other VoteFragment) bool {
	return v.LastZxid.Equal(other.LastZxid) && v.Id == other.Id
}

func (v VoteFragment) GreaterThan(other VoteFragment) bool {
	return v.LastZxid.GreaterThan(other.LastZxid) || (v.LastZxid.Equal(other.LastZxid) && v.Id > other.Id)
}
