package types

type Event struct {
	ID     string
	Source string
	Type   string
	Action string

	Actor  string
	Target string

	Payload map[string]any

	Timestamp int64
}
