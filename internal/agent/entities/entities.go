package entities

type ParsedRequest struct {
	Command Command
	Key     string
	Data    string
}

type Command int

const (
	Unknown Command = iota
	GET
	SET
	DEL
	//LISTNODES
)
