package zouk

type VoteFragment struct {
	LastZxid ZxidFragment
	Id       int
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
