package zouk

func (v *Vote) LessThan(other *Vote) bool {
	return v.LastZxid.LessThan(other.LastZxid) || (v.LastZxid.Equal(other.LastZxid) && v.Id < other.Id)
}

func (v *Vote) Equal(other *Vote) bool {
	return v.LastZxid.Equal(other.LastZxid) && v.Id == other.Id
}

func (v *Vote) GreaterThan(other *Vote) bool {
	return v.LastZxid.GreaterThan(other.LastZxid) || (v.LastZxid.Equal(other.LastZxid) && v.Id > other.Id)
}
