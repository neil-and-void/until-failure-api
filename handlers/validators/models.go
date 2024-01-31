package validators

// *** Requests ***
type (
	User struct {
		ID string `json:"id" validate:"required"`
	}

	Routine struct {
		ID string `json`
	}
)

// *** Responses ***
type (
	ErrorResponse struct {
		Error string `json:"Error"`
	}
)
