package pkg

var (
	SUCCESS      = NewResultCode(200, "Success")
	FAIL         = NewResultCode(100, "FAIL")
	NOT_FOUNT    = NewResultCode(404, "Not Found")
	SYSTEM_ERROR = NewResultCode(500, "System Error")
)

type resultCode struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewResultCode(code int, msg string) *resultCode {
	return &resultCode{
		Code: code,
		Msg:  msg,
	}
}
