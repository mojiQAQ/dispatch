package model

type (
	ReqBase struct {
		RequestID string `json:"request_uuid"`
		Cookie    string `json:"cookie"`
	}

	RespBase struct {
		Message string `json:"message"`
	}

	Page struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}

	PageInfo struct {
		*Page
		Total int `json:"total"`
	}
)

func (r *ReqBase) GenResponse(err error) *RespBase {

	msg := "ok"
	if err != nil {
		msg = err.Error()
	}

	res := &RespBase{
		Message: msg,
	}

	return res
}
