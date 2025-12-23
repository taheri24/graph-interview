package handlers

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// NewErrorResponse creates a new ErrorResponse with the given error and optional message
func NewErrorResponse(error string, message ...string) ErrorResponse {
	resp := ErrorResponse{Error: error}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	return resp
}

// NewErr creates a new ErrorResponse with the given error(err) and optional message
func NewErr(err error, message ...string) ErrorResponse {
	if err == nil {
		panic("you never call `NewErr` with err=nil")
	}

	var errorMsg string = err.Error()
	return NewErrorResponse(errorMsg, message...)
}
