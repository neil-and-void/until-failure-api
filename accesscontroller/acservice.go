package accesscontroller

type AccessControllerService interface {
	CanAccessWorkoutRoutine(userId string, workoutRoutineId string) error
	CanAccessWorkoutSession(userId string, workoutSessionId string) error
}
