package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"github.com/neilZon/workout-logger-api/graph/generated"
)

// Exercise returns generated.ExerciseResolver implementation.
func (r *Resolver) Exercise() generated.ExerciseResolver { return &exerciseResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// PrevExercise returns generated.PrevExerciseResolver implementation.
func (r *Resolver) PrevExercise() generated.PrevExerciseResolver { return &prevExerciseResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// WorkoutRoutine returns generated.WorkoutRoutineResolver implementation.
func (r *Resolver) WorkoutRoutine() generated.WorkoutRoutineResolver {
	return &workoutRoutineResolver{r}
}

// WorkoutSession returns generated.WorkoutSessionResolver implementation.
func (r *Resolver) WorkoutSession() generated.WorkoutSessionResolver {
	return &workoutSessionResolver{r}
}

type exerciseResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type prevExerciseResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type workoutRoutineResolver struct{ *Resolver }
type workoutSessionResolver struct{ *Resolver }
