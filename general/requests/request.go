package requests

// Request .
type Request interface {
	Post() Request
	Put() Request
	Get() Request
	Delete() Request
	SetJSONBody(obj interface{}) Request
	SetBody(byts []byte) Request
	ToJSON(obj interface{}, httpRespones ...*Response) error
	ToString(httpRespones ...*Response) (string, error)
	ToBytes(httpRespones ...*Response) ([]byte, error)
	ToXML(v interface{}, httpRespones ...*Response) (e error)
	SetHeader(key, value string) Request
}

// Response .
type Response struct {
	Header        map[string]string
	ContentLength int64
	ContentType   string
	StatusCode    int
	HTTP11        bool
}
