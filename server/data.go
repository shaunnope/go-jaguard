package main

import (
	"log/slog"
	"os"

	"github.com/shaunnope/go-jaguard/utils"
	pb "github.com/shaunnope/go-jaguard/zouk"
)

const (
	data_STATE   = "state"   // vote (can deduce state from vote)
	data_EPOCH   = "epoch"   // accepted and current
	data_HISTDIR = "history" // history directory
	data_ZXID    = "zxid"    // lastZxid
)

// Load the state of the server from persistent storage
func (s *StateVector) LoadStates() error {
	if err := os.MkdirAll(s.Path, 0755); err != nil {
		return err
	}

	// load vote and state
	if data, err := os.ReadFile(s.Path + data_STATE); err != nil {
		// slog.Error("LoadStates Vote", "err", err)
		return err
	} else {
		if err := s.Vote.Unmarshal(data); err != nil {
			// slog.Error("LoadStates Vote", "err", err)
			return err
		}
		if s.Vote.Id != s.Id {
			s.State = FOLLOWING
		} else {
			s.State = LEADING
		}
	}

	// load epoch
	if data, err := os.ReadFile(s.Path + data_EPOCH); err != nil {
		// slog.Error("LoadStates Epoch", "err", err)
		return err
	} else {
		if len(data) != 16 {
			// slog.Error("LoadStates Epoch", "err", "invalid epoch length")
			return err
		}
		s.AcceptedEpoch = utils.UnmarshalInt(data[0:8])
		s.CurrentEpoch = utils.UnmarshalInt(data[8:16])
	}

	// load zxid
	if data, err := os.ReadFile(s.Path + data_ZXID); err != nil {
		// slog.Error("LoadStates Zxid", "err", err)
		return err
	} else {
		s.LastZxid.Unmarshal(data)
	}

	// load history
	if err := os.MkdirAll(s.Path+data_HISTDIR, 0755); err != nil {
		// slog.Error("LoadStates History", "err", err)
		return err
	}
	files, err := os.ReadDir(s.Path + data_HISTDIR)
	if err != nil {
		// slog.Error	("LoadStates History", "err", err)
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := file.Name()
		data, err := os.ReadFile(s.Path + data_HISTDIR + "/" + path)
		if err != nil {
			slog.Error("LoadStates History", "path", path, "err", err)
			continue
		}
		if len(data) < 24 {
			slog.Error("LoadStates History", "path", path, "err", "invalid data length")
			continue
		}
		isEphemeral := data[0] == 1
		isSequential := data[1] == 1
		zxid := pb.ZxidFragment{Epoch: utils.UnmarshalInt(data[8:16]), Counter: utils.UnmarshalInt(data[16:24])}
		data = data[24:]
		s.Data.CreateNode(path, data, isEphemeral, 1, zxid, isSequential)
	}

	return nil
}

func (s *StateVector) SaveState(name string, data []byte) error {
	// write data to file
	return utils.WriteBytes(data, s.Path, name)
}

func (s *StateVector) SaveHistory() error {
	return nil
}
