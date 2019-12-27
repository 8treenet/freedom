package infra

import (
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/hero"
)

type Response struct {
	Code        int
	ContentType string
	Content     []byte

	// if not empty then content type is the text/plain
	// and content is the text as []byte.
	Text string
	// If not nil then it will fire that as "application/json" or the
	// "ContentType" if not empty.
	Object interface{}

	// If Path is not empty then it will redirect
	// the client to this Path, if Code is >= 300 and < 400
	// then it will use that Code to do the redirection, otherwise
	// StatusFound(302) or StatusSeeOther(303) for post methods will be used.
	// Except when err != nil.
	Path string

	// if not empty then fire a 400 bad request error
	// unless the Status is > 200, then fire that error code
	// with the Err.Error() string as its content.
	//
	// if Err.Error() is empty then it fires the custom error handler
	// if any otherwise the framework sends the default http error text based on the status.
	Err error
	Try func() int

	// if true then it skips everything else and it throws a 404 not found error.
	// Can be named as Failure but NotFound is more precise name in order
	// to be visible that it's different than the `Err`
	// because it throws a 404 not found instead of a 400 bad request.
	// NotFound bool
	// let's don't add this yet, it has its dangerous of missuse.

	ErrCode int
	ErrMsg  string
}

func (r Response) Dispatch(ctx context.Context) {
	if r.Path != "" && r.Err == nil {
		// it's not a redirect valid status
		if r.Code < 300 || r.Code >= 400 {
			if ctx.Method() == "POST" {
				r.Code = 303 // StatusSeeOther
			}
			r.Code = 302 // StatusFound
		}
		ctx.Redirect(r.Path, r.Code)
		return
	}

	if s := r.Text; s != "" {
		r.Content = []byte(s)
	}
	var obj struct {
		ErrCode int         `json:"errcode"`
		ErrMsg  string      `json:"errmsg"`
		Data    interface{} `json:"data,omitempty"`
	}
	obj.Data = r.Object
	obj.ErrCode = r.ErrCode
	obj.ErrMsg = r.ErrMsg

	hero.DispatchCommon(ctx, r.Code, r.ContentType, r.Content, obj, r.Err, true)
}
