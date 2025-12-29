package main

type datatype int8

const (
	DATATYPE_ERR datatype = iota - 1
	DATATYPE_MISC
	DATATYPE_OPCODE
	DATATYPE_STRING
	DATATYPE_HEX
	DATATYPE_NUMBER
	DATATYPE_TIME
	DATATYPE_ADDRESS
)

func (d datatype) String() string {
	switch d {
	case DATATYPE_ERR:
		return "err"
	case DATATYPE_OPCODE:
		return "opcode"
	case DATATYPE_MISC:
		return "misc"
	case DATATYPE_STRING:
		return "string"
	case DATATYPE_HEX:
		return "hex"
	case DATATYPE_NUMBER:
		return "number"
	case DATATYPE_TIME:
		return "time"
	case DATATYPE_ADDRESS:
		return "address"
	default:
		return "unknown"
	}
}

func (d datatype) Color() string {
	switch d {
	case DATATYPE_TIME:
		fallthrough
	case DATATYPE_NUMBER:
		return "blue"
	case DATATYPE_HEX:
		fallthrough
	case DATATYPE_STRING:
		return "green"
	case DATATYPE_MISC:
		return "magenta"
	case DATATYPE_OPCODE:
		return "yellow"
	case DATATYPE_ERR:
		return "red"
	default:
		return "whitebright"
	}
}
