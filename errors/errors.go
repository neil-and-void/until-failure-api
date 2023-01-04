package errors

// package contains error strings that are expected to be presented to the user
const (
	CreateWorkoutRoutineError = "Could not create workout routine, %s"
	GetWorkoutRoutinesError   = "Could not get workout routines, %s"
	GetWorkoutRoutineError    = "Could not get workout routine, %s"
	UpdateWorkoutRoutineError = "Could not update workout routine, %s"
	DeleteWorkoutRoutineError = "Could not delete workout routine, %s"

	CreateWorkoutSessionError = "Could not create workout session, %s"
	GetWorkoutSessionsError   = "Could not get workout sessions, %s"
	GetWorkoutSessionError    = "Could not get workout session, %s"
	UpdateWorkoutSessionError = "Could not update session session, %s"
	DeleteWorkoutSessionError = "Could not delete workout session, %s"
)
