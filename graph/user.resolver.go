package graph

import (
	"context"
	"fmt"

	"github.com/neilZon/workout-logger-api/graph/model"
)

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context) (int, error) {
	panic(fmt.Errorf("not implemented: DeleteUser - deleteUser"))
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented: User - user"))
}
