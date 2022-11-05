package accesscontroller

// need to put this in a separate package from accesscontrol to prevent circular import
type AccessControllerService interface {
	CanAccessWorkoutRoutine(userId string, workoutRoutineId string) error
	CanAccessWorkoutSession(userId string, workoutSessionId string) error
}
