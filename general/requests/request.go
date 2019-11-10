package requests

// Request .
type Request interface {
	Post() Request
	Put() Request
	Get() Request
	Delete() Request
	SetJSONBody(obj interface{}) Request
	SetBody(byts []byte) Request
	//跳过总线数据,上下游header携带
	SkipBus() Request
	ToJSON(obj interface{}) Response
	ToString() (string, Response)
	ToBytes() ([]byte, Response)
	ToXML(v interface{}) Response
	SetHeader(key, value string) Request
	SetParam(key string, value interface{}) Request
	URI() string
}

// Response .
type Response struct {
	Error         error
	Header        map[string]string
	ContentLength int64
	ContentType   string
	StatusCode    int
	HTTP11        bool
}
