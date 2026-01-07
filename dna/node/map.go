package node

type Flag string

const (
	Root                  Flag = "root"
	Idle                       = "Idle"
	Task                       = "Task"
	Sleep                      = "Sleep"
	Watch                      = "Watch"
	Explore                    = "Explore"
	Introspection              = "Introspection"
	Answer                     = "Answer"
	Memory                     = "Memory"
	Ask                        = "Ask"
	WatchEnvironment           = "WatchEnvironment"
	WatchRoom                  = "WatchRoom"
	ExploreConcept             = "ExploreConcept"
	ExploreBehavior            = "ExploreBehavior"
	ExploreCharacter           = "ExploreCharacter"
	OrganizeMetacognition      = "OrganizeMetacognition"
)
