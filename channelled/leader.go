package main

import (
	"fmt"
	"sync"
	"time"
)

type OperationType int

const (
	CREATE OperationType = iota
	UPDATE
	DELETE
)

type Transaction struct {
	ID    int
	Type  OperationType
	Key   string
	Value string
}

type ResponseType int

const (
	AGREE ResponseType = iota
	ABORT
)

type FollowerResponse struct {
	follower *Follower
	response ResponseType
}

type Leader struct {
	State                map[string]string
	TransactionLog       []Transaction
	CurrentTransactionID int
	Mutex                sync.Mutex
	Followers            []*Follower
	ResponseChannel      chan FollowerResponse
}

type Follower struct {
	ID    int
	State map[string]string
}

func NewLeader() *Leader {
	return &Leader{
		State:                make(map[string]string),
		TransactionLog:       make([]Transaction, 0),
		CurrentTransactionID: 0,
		ResponseChannel:      make(chan FollowerResponse),
	}
}

func (l *Leader) AddFollower(f *Follower) {
	l.Followers = append(l.Followers, f)
}

func (l *Leader) HandleWriteRequest(opType OperationType, key, value string) error {
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	switch opType {
	case CREATE:
		if _, exists := l.State[key]; !exists {
			l.State[key] = value
			l.recordTransaction(CREATE, key, value)
			return nil
		} else {
			return fmt.Errorf("Key already exists")
		}
	case UPDATE:
		if _, exists := l.State[key]; exists {
			l.State[key] = value
			l.recordTransaction(UPDATE, key, value)
			return nil
		} else {
			return fmt.Errorf("Key not found")
		}
	case DELETE:
		if _, exists := l.State[key]; exists {
			delete(l.State, key)
			l.recordTransaction(DELETE, key, "")
			return nil
		} else {
			return fmt.Errorf("Key not found")
		}
	}
	return fmt.Errorf("Invalid operation type")
}

func (l *Leader) recordTransaction(opType OperationType, key, value string) {
	transaction := Transaction{
		ID:    l.CurrentTransactionID + 1,
		Type:  opType,
		Key:   key,
		Value: value,
	}
	if l.propagateTransactionToFollowers(transaction) {
		l.CurrentTransactionID++
		l.TransactionLog = append(l.TransactionLog, transaction)
	} else {
		fmt.Println("Transaction aborted by 2PC")
	}
}

func (l *Leader) propagateTransactionToFollowers(transaction Transaction) bool {
	for _, follower := range l.Followers {
		go l.sendPrepareMessage(follower, transaction)
	}

	votesReceived := 0
	abortFlag := false

	timeout := time.After(10 * time.Second)
	for i := 0; i < len(l.Followers); i++ {
		select {
		case response := <-l.ResponseChannel:
			if response.response == AGREE {
				votesReceived++
			} else if response.response == ABORT {
				abortFlag = true
			}
		case <-timeout:
			fmt.Println("Timeout waiting for follower response")
			abortFlag = true
		}
	}

	if votesReceived == len(l.Followers) && !abortFlag {
		for _, follower := range l.Followers {
			go l.sendCommitMessage(follower)
		}
		return true
	} else {
		for _, follower := range l.Followers {
			go l.sendAbortMessage(follower)
		}
		return false
	}
}

func (l *Leader) sendPrepareMessage(follower *Follower, transaction Transaction) {
	// Simulating the follower's decision. For simplicity, we're always agreeing.
	// In reality, more complex logic may decide whether to AGREE or ABORT.
	response := AGREE
	l.ResponseChannel <- FollowerResponse{follower: follower, response: response}
}

func (l *Leader) sendCommitMessage(follower *Follower) {
	// In a real system, we would send a commit message to the follower and await acknowledgment.
}

func (l *Leader) sendAbortMessage(follower *Follower) {
	// In a real system, we would send an abort message to the follower.
}

func main() {
	leader := NewLeader()

	// Adding some followers
	follower1 := &Follower{ID: 1, State: make(map[string]string)}
	follower2 := &Follower{ID: 2, State: make(map[string]string)}
	leader.AddFollower(follower1)
	leader.AddFollower(follower2)

	// Test CREATE operation
	err := leader.HandleWriteRequest(CREATE, "testKey", "testValue")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("CREATE operation successful.")
	}

	// Print state and transaction log
	fmt.Println("State:", leader.State)
	fmt.Println("Transaction Log:", leader.TransactionLog)
}
