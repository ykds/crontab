package pkg

import "encoding/json"

type result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (r *result) WithMsg(msg string) *result {
	r.Msg = msg
	return r
}

func (r *result) WithData(data interface{}) *result {
	r.Data = data
	return r
}

func (r *result) ToJson() []byte {
	data, _ := json.Marshal(r)
	return data
}

func Success() *result {
	return &result{
		Code: SUCCESS.Code,
		Msg:  SUCCESS.Msg,
	}
}

func NotFound(msg string) *result {
	return &result{
		Code: NOT_FOUNT.Code,
		Msg:  msg,
	}
}

func Fail(msg string) *result {
	return &result{
		Code: FAIL.Code,
		Msg:  msg,
	}
}

func SystemError(msg string) *result {
	return &result{
		Code: SYSTEM_ERROR.Code,
		Msg:  msg,
	}
}
