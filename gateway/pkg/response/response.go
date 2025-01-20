package response

// Response adalah struktur standar untuk response API
type Response struct {
	Status     bool        `json:"status"`
	StatusCode int32       `json:"statusCode"`
	Message    *string     `json:"message,omitempty"`
	Error      *string     `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func NewResponse() *Response {
	return &Response{
		Status:     true,
		StatusCode: 200,
	}
}

func Success(data interface{}, message string) *Response {
	return &Response{
		Status:     true,
		StatusCode: 200,
		Message:    &message,
		Data:       data,
	}
}

func Error(statusCode int32, errorMessage string) *Response {
	return &Response{
		Status:     false,
		StatusCode: statusCode,
		Error:      &errorMessage,
	}
}

func (r *Response) WithData(data interface{}) *Response {
	r.Data = data
	return r
}

func (r *Response) WithMessage(message string) *Response {
	r.Message = &message
	return r
}

func (r *Response) WithError(err string) *Response {
	r.Error = &err
	r.Status = false
	return r
}

// WithStatusCode mengubah status code response
func (r *Response) WithStatusCode(code int32) *Response {
	r.StatusCode = code
	return r
}
