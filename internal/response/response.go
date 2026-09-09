package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Response tra ve JSON chuan khi thanh cong (Code = 0)
func Response(w http.ResponseWriter, resp interface{}) {
	body := Body{
		Code: 0,
		Msg:  "success",
		Data: resp,
	}
	httpx.OkJson(w, body)
}

// Error tra ve JSON chuan khi co loi (Code = -1)
func Error(w http.ResponseWriter, err error) {
	body := Body{
		Code: -1,
		Msg:  err.Error(),
		Data: nil,
	}
	httpx.WriteJson(w, http.StatusOK, body)
}