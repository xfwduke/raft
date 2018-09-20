package main

type FSMState int

const (
	Follower FSMState = iota
	Candidate
	Leader
)

type FSMEvent int

const (
	ElectionTimeout FSMEvent = iota
	RecvVoteRequest
	RecvAppendEntriesRequest
)

type FSM struct {
	currentState FSMState
}

func main() {

}
