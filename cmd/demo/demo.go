package main

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (r NodeState) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[r]
}

type Node struct {
	currentTerm uint64
	votedFor    uint64
	nodeState NodeState
}


func main() {
	
}
