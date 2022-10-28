package leader_checker

type Checker interface {
	IsLeader() bool
}
