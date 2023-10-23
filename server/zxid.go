package main

import (
	pb "github.com/shaunnope/go-jaguard/zouk"
)

type Zxid struct {
	Epoch   int
	Counter int
}

func (z *Zxid) LessThan(other Zxid) bool {
	return z.Epoch < other.Epoch || (z.Epoch == other.Epoch && z.Counter < other.Counter)
}

func (z *Zxid) Equal(other Zxid) bool {
	return z.Epoch == other.Epoch && z.Counter == other.Counter
}

func (z *Zxid) GreaterThan(other Zxid) bool {
	return z.Epoch > other.Epoch || (z.Epoch == other.Epoch && z.Counter > other.Counter)
}

func ZxidFrom(raw *pb.Zxid) *Zxid {
	return &Zxid{Epoch: int(raw.Epoch), Counter: int(raw.Counter)}
}

func (z *Zxid) Raw() *pb.Zxid {
	return &pb.Zxid{Epoch: int64(z.Epoch), Counter: int64(z.Counter)}
}
