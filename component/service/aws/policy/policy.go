package policy

func New(s ...Statement) Policy {
	return Policy{
		Version:   "2012-10-17",
		Statement: s,
	}
}

type Policy struct {
	Version   string      `json:"Version,omitempty"`
	Statement []Statement `json:"Statement,omitempty"`
}

type Statement struct {
	Sid       string      `json:"Sid,omitempty"`
	Effect    string      `json:"Effect,omitempty"`
	Principal interface{} `json:"Principal,omitempty"`
	Action    []string    `json:"Action,omitempty"`
	Resource  interface{} `json:"Resource,omitempty"`
	Condition *Condition  `json:"Condition,omitempty"`
}

type Principal struct {
	AWS interface{} `json:"AWS,omitempty"`
}

type Condition struct {
	StringNotLike map[string]interface{} `json:"StringNotLike,omitempty"`
	StringLike    map[string]interface{} `json:"StringLike,omitempty"`
}
