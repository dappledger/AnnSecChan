package api

type ReqPut struct {
	PubKeys []string `json:"pubkeys"`
	Value   []byte   `json:"value"`
}

type ReqGet struct {
	Key []byte `json:"key"`
}
