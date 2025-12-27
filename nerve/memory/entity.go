package memory

type Entity struct {
	Id          string `json:"id"`
	Subject     string `json:"subject"`
	SubjectType string `json:"subject_type"`
	Predicate   string `json:"predicate"`
	Object      string `json:"object"`
	ObjectType  string `json:"object_type"`
	// Link is the link to the enduring memory.
	Link string `json:"link"`
}
