package graph

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// AddWorkoutSession is the resolver for the addWorkoutSession field.
func (r *mutationResolver) AddWorkoutSession(ctx context.Context, workout model.WorkoutSessionInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	var dbExercises []database.Exercise
	for _, e := range workout.Exercises {
		var set []database.SetEntry

		for _, s := range e.SetEntries {
			set = append(set, database.SetEntry{
				Weight: float32(s.Weight),
				Reps:   uint(s.Reps),
			})
		}

		exerciseRoutineId, err := strconv.ParseUint(e.ExerciseRoutineID, 10, 32)
		if err != nil {
			return "", gqlerror.Errorf("Error Adding Workout Session")
		}

		dbExercises = append(dbExercises, database.Exercise{
			Sets:              set,
			ExerciseRoutineID: uint(exerciseRoutineId),
			Notes:             e.Notes,
		})
	}

	workotuRoutineID, err := strconv.ParseUint(workout.WorkoutRoutineID, 10, 64)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Workout Session: Invalid Workout Routine ID")
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
		return "", gqlerror.Errorf("Error Adding Workout Session")
	}

	return fmt.Sprintf("%d", ws.ID), nil
}

// UpdateWorkoutSession is the resolver for the updateWorkoutSession field.
func (r *mutationResolver) UpdateWorkoutSession(ctx context.Context, workoutSessionID string, updateWorkoutSessionInput model.UpdateWorkoutSessionInput) (*model.UpdatedWorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.UpdatedWorkoutSession{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return &model.UpdatedWorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session: Access Denied")
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
		return &model.UpdatedWorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session")
	}

	return &model.UpdatedWorkoutSession{
		ID:    fmt.Sprintf("%d", updatedWorkoutSession.ID),
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

	userId := fmt.Sprintf("%d", u.ID)
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
func (r *queryResolver) WorkoutSessions(ctx context.Context) ([]*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.WorkoutSession{}, err
	}

	dbWorkoutSessions, err := database.GetWorkoutSessions(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return []*model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Sessions")
	}

	var workoutSessions []*model.WorkoutSession
	for _, ws := range dbWorkoutSessions {

		workoutSession := &model.WorkoutSession{
			ID:               fmt.Sprintf("%d", ws.ID),
			Start:            ws.Start,
			End:              ws.End,
			WorkoutRoutineID: fmt.Sprintf("%d", ws.WorkoutRoutineID),
		}

		exercises, err := r.Resolver.WorkoutSession().Exercises(ctx, workoutSession)
		if err != nil {
			return []*model.WorkoutSession{}, nil
		}
		workoutSession.Exercises = exercises

		workoutSessions = append(workoutSessions, workoutSession)
	}

	return workoutSessions, nil
}

// WorkoutSession is the resolver for the workoutSession field.
func (r *queryResolver) WorkoutSession(ctx context.Context, workoutSessionID string) (*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	var dbWorkoutSession database.WorkoutSession
	err = database.GetWorkoutSession(r.DB, fmt.Sprintf("%d", u.ID), workoutSessionID, &dbWorkoutSession)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Session: Access Denied")
	}

	exercises, err := r.Resolver.WorkoutSession().Exercises(ctx, &model.WorkoutSession{
		ID: workoutSessionID,
	})
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	return &model.WorkoutSession{
		ID:               fmt.Sprintf("%d", dbWorkoutSession.ID),
		Start:            dbWorkoutSession.Start,
		End:              dbWorkoutSession.End,
		WorkoutRoutineID: fmt.Sprintf("%d", dbWorkoutSession.WorkoutRoutineID),
		Exercises:        exercises,
	}, nil
}
