package graph

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/errors"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// AddWorkoutSession is the resolver for the addWorkoutSession field.
func (r *mutationResolver) AddWorkoutSession(ctx context.Context, workout model.WorkoutSessionInput) (*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	var dbExercises []database.Exercise
	for _, e := range workout.Exercises {
		var set []database.SetEntry

		for _, s := range e.SetEntries {
			var weight *float32
			if s.Weight != nil {
				temp := float32(*s.Weight)
				weight = &temp
			}
			var reps *uint
			if s.Reps != nil {
				temp := uint(*s.Reps)
				reps = &temp
			}

			set = append(set, database.SetEntry{
				Weight: weight,
				Reps:   reps,
			})
		}

		exerciseRoutineId, err := strconv.ParseUint(e.ExerciseRoutineID, 10, 32)
		if err != nil {
			return &model.WorkoutSession{}, gqlerror.Errorf("Error Adding Workout Session")
		}

		dbExercises = append(dbExercises, database.Exercise{
			Sets:              set,
			ExerciseRoutineID: uint(exerciseRoutineId),
			Notes:             e.Notes,
		})
	}

	workotuRoutineID, err := strconv.ParseUint(workout.WorkoutRoutineID, 10, 64)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Adding Workout Session: Invalid Workout Routine ID")
	}

	ws := &database.WorkoutSession{
		Start:            workout.Start,
		End:              workout.End,
		WorkoutRoutineID: uint(workotuRoutineID),
		UserID:           u.ID,
		Exercises:        dbExercises,
	}
	err = database.AddWorkoutSession(r.DB, ws)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Adding Workout Session")
	}

	return &model.WorkoutSession{
		ID: utils.UIntToString(ws.ID),
		// return so previous exercise routine resolver can use
		WorkoutRoutine: model.WorkoutRoutine{
			ID: workout.WorkoutRoutineID,
		},
		Start: ws.Start,
		End:   ws.End,
	}, nil
}

// UpdateWorkoutSession is the resolver for the updateWorkoutSession field.
func (r *mutationResolver) UpdateWorkoutSession(ctx context.Context, workoutSessionID string, updateWorkoutSessionInput model.UpdateWorkoutSessionInput) (*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	userId := utils.UIntToString(u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session: Access Denied")
	}

	var start time.Time
	if updateWorkoutSessionInput.Start != nil {
		start = *updateWorkoutSessionInput.Start
	}
	updatedWorkoutSession := database.WorkoutSession{
		Start: start,
		End:   updateWorkoutSessionInput.End,
	}
	err = database.UpdateWorkoutSession(r.DB, workoutSessionID, &updatedWorkoutSession)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session")
	}

	return &model.WorkoutSession{
		ID:    utils.UIntToString(updatedWorkoutSession.ID),
		Start: updatedWorkoutSession.Start,
		End:   updatedWorkoutSession.End,
	}, nil
}

// DeleteWorkoutSession is the resolver for the deleteWorkoutSession field.
func (r *mutationResolver) DeleteWorkoutSession(ctx context.Context, workoutSessionID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return 0, err
	}

	userId := utils.UIntToString(u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Session: Access Denied")
	}

	err = database.DeleteWorkoutSession(r.DB, workoutSessionID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Session")
	}

	return 1, nil
}

// WorkoutSessions is the resolver for the workoutSessions field.
func (r *queryResolver) WorkoutSessions(ctx context.Context, limit int, after *string) (*model.WorkoutSessionConnection, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSessionConnection{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutSessionConnection{}, err
	}

	if limit <= 0 || limit > 30 {
		return &model.WorkoutSessionConnection{}, gqlerror.Errorf(errors.GetWorkoutRoutinesError, "limit needs to be between 1 to 30")
	}

	cursor := ""
	if after != nil && *after != "" {
		cursor = *after
	}

	dbWorkoutSessions, err := database.GetWorkoutSessions(r.DB, utils.UIntToString(u.ID), cursor, limit)
	if err != nil {
		return &model.WorkoutSessionConnection{}, gqlerror.Errorf(errors.GetWorkoutSessionsError)
	}

	var edges []*model.WorkoutSessionEdge
	for _, workoutSession := range dbWorkoutSessions {
		edges = append(edges, &model.WorkoutSessionEdge{
			Cursor: utils.UIntToString(workoutSession.ID),
			Node: &model.WorkoutSession{
				ID: utils.UIntToString(workoutSession.ID),
				// return workout routine to access in exercise resolver
				WorkoutRoutine: model.WorkoutRoutine{
					ID: utils.UIntToString(workoutSession.WorkoutRoutineID),
				},
				Start: workoutSession.Start,
				End:   workoutSession.End,
			},
		})
	}

	return &model.WorkoutSessionConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage: true,
		},
	}, nil
}

// WorkoutSession is the resolver for the workoutSession field.
func (r *queryResolver) WorkoutSession(ctx context.Context, workoutSessionID string) (*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	workoutSession, err := database.GetUsersWorkoutSession(r.DB, workoutSessionID, utils.UIntToString(u.ID))
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Session: Access Denied")
	}

	return &model.WorkoutSession{
		ID: utils.UIntToString(workoutSession.ID),
		// return workout routine ID to access in workout routine resolver
		WorkoutRoutine: model.WorkoutRoutine{
			ID: utils.UIntToString(workoutSession.WorkoutRoutineID),
		},
		Start: workoutSession.Start,
		End:   workoutSession.End,
	}, nil
}
