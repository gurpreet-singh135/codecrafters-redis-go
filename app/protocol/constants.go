package protocol

const (
	CRLF              = "\r\n"
	RESPONSE_OK       = "OK"
	RESPONSE_NONE     = "none"
	INVALID_ENTRY_ID  = "The ID specified in XADD is equal or smaller than the target stream top item"
	INVALID_MIN_ID    = "The ID specified in XADD must be greater than 0-0"
	NOT_AN_INTEGER    = "value is not an integer or out of range"
	EMPTY_STRING      = ""
	BLOCK_STRING      = "BLOCK"
	EXEC_BEFORE_MULTI = "EXEC without MULTI"
)

var (
	RESP_ZERO_REQUEST = []string{}
)
