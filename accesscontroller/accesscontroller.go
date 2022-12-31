package accesscontroller

// need to put this in a separate package from accesscontrol to prevent circular import
type AccessControllerService interface {
	CanAccessWorkoutRoutine(userId string, workoutRoutineId string) error
	CanAccessWorkoutSession(userId string, workoutSessionId string) error
	CanAccessExerciseRoutine(userId string, exerciseId string) error
	CanAccessExercise(userId string, exerciseId string) error
	CanAccessSetEntry(userId string, exerciseId string) error
}
