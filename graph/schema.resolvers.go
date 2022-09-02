package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email *string, password *string) (model.AuthResult, error) {
	panic(fmt.Errorf("not implemented: Login - login"))
}

// Signup is the resolver for the signup field.
func (r *mutationResolver) Signup(ctx context.Context, email *string, name *string, password *string, confirmPassword *string) (model.AuthResult, error) {
	fmt.Println(*email, *name, *password, *confirmPassword)

	if *password != *confirmPassword {
		return nil, gqlerror.Errorf("Passwords don't match")
	}

	if _, err := mail.ParseAddress(*email); err == nil {
		return nil, gqlerror.Errorf("Not a valid email")
	}

	panic(fmt.Errorf("not implemented: Signup - signup"))
}

// WorkoutRoutines is the resolver for the workoutRoutines field.
func (r *queryResolver) WorkoutRoutines(ctx context.Context) ([]*model.WorkoutRoutine, error) {
	panic(fmt.Errorf("not implemented: WorkoutRoutines - workoutRoutines"))
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *queryResolver) ExerciseRoutines(ctx context.Context) ([]*model.ExerciseRoutine, error) {
	panic(fmt.Errorf("not implemented: ExerciseRoutines - exerciseRoutines"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
