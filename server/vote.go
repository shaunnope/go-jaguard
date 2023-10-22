// Vote related definitions
package main

type Vote [2]int

type VoteLog struct {
	vote    Vote
	round   int
	version int
}

type VoteMsg struct {
	vote  Vote
	id    int
	state State
	round int
}

func (v *Vote) LessThan(other Vote) bool {
	return v[0] < other[0] || (v[0] == other[0] && v[1] < other[1])
}

func (v *Vote) Equal(other Vote) bool {
	return v[0] == other[0] && v[1] == other[1]
}

func (v *Vote) GreaterThan(other Vote) bool {
	return v[0] > other[0] || (v[0] == other[0] && v[1] > other[1])
}
