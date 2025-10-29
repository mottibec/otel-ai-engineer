package humanaction

// RespondToHumanActionRequest represents the request to respond to a human action
type RespondToHumanActionRequest struct {
	Response string `json:"response"`
	Resume   bool   `json:"resume,omitempty"` // Whether to immediately resume the agent run
}

