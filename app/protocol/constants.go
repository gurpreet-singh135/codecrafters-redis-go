package protocol

const (
	CRLF         = "\r\n"
	RESPONSE_OK  = "OK"
	RESPONSE_NONE = "none"
	INVALID_ENTRY_ID = "The ID specified in XADD is equal or smaller than the target stream top item"
	INVALID_MIN_ID = "The ID specified in XADD must be greater than 0-0"
)

var (
	RESP_ZERO_REQUEST = []string{}
)