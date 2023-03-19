package graph

import (
	"context"
	"fmt"

	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return 0, err
	}

	err = database.DeleteUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return 0, err
	}
	return 1, err
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.User{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.User{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	user, err := database.GetUserById(r.DB, userId)
	if err != nil {
		return &model.User{}, err
	}
	if user == nil {
		return &model.User{}, gqlerror.Errorf("User does not exist")
	}

	return &model.User{
		ID:    userId,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}
