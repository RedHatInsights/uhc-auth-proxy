package client

type HttpError struct {
	Message    string
	StatusCode int
}

func (h *HttpError) Error() string {
	return h.Message
}
