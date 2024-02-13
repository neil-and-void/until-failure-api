package handlers

// *** Responses ***
type (
	ErrorResponse struct {
		Error string `json:"Error"`
	}

	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	Routine struct {
		ID               string            `json:"id"`
		Name             string            `json:"name"`
		ExerciseRoutines []ExerciseRoutine `json:"exerciseRoutines"`
		Active           bool              `json:"active"`
		Private          bool              `json:"private"`
		UserID           string            `json:"userId"`
		CreatedAt        string            `json:"createdAt"`
	}

	ExerciseRoutine struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Active    bool   `json:"active"`
		RoutineId string `json:"routineId"`
		CreatedAt string `json:"createdAt"`
	}

	// Workout
	// Exercise
	// Set
)
