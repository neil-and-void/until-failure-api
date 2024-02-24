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
		ID         string      `json:"id"`
		Name       string      `json:"name"`
		Active     bool        `json:"active"`
		RoutineId  string      `json:"routineId"`
		SetSchemes []SetScheme `json:"setSchemes"`
		CreatedAt  string      `json:"createdAt"`
	}

	SetScheme struct {
		ID                string          `json:"id"`
		TargetReps        uint            `json:"targetReps"`
		SetType           SetType         `json:"setType"`
		Measurement       MeasurementType `json:"measurement"`
		ExerciseRoutineId string          `json:"exerciseRoutineId"`
		CreatedAt         string          `json:"createdAt"`
	}

	// Workout
	// Exercise
	// Set
)
