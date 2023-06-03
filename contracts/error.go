package contracts

type HttpError struct {
	Message    string `json:"message"`
	IncidentID string `json:"incident_id,omitempty"`
}
