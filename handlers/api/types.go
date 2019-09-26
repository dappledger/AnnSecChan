package api

type ReqSignPut struct {
	PubKeys []string `json:"public_keys"`
	Value   []byte   `json:"value"`
	Sign    []byte   `json:"sign"`
}

type ReqPut struct {
	PubKeys []string `json:"public_keys"`
	Value   []byte   `json:"value"`
}

type ReqGet struct {
	Key []byte `json:"key"`
}

type ReqStartRecoverTask struct {
	PubKey   string `json:"public_key"`
	IsResume bool   `json:"is_resume"`
}
