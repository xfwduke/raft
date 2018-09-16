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
	nodeState   NodeState
}

func (c *Node) Run() {
	c.transToFollower()
}

func (c *Node) transToFollower() {
	c.nodeState = Follower
}
func (c *Node) transToCandidate() {
	c.nodeState = Candidate
}
func (c *Node) transToLeader() {
	c.nodeState = Leader
}

func main() {

}
