package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

// Sets is the resolver for the sets field.
func (r *exerciseResolver) Sets(ctx context.Context, obj *model.Exercise) ([]*model.SetEntry, error) {
	var dbSetEntries []database.SetEntry
	err := database.GetSets(r.DB, &dbSetEntries, obj.ID)
	if err != nil {
		return []*model.SetEntry{}, nil
	}

	var setEntries []*model.SetEntry
	for _, s := range dbSetEntries {
		setEntries = append(setEntries, &model.SetEntry{
			ID:     fmt.Sprintf("%d", s.ID),
			Weight: float64(s.Weight),
			Reps:   int(s.Reps),
		})
	}

	return setEntries, nil
}

// AddSet is the resolver for the addSet field.
func (r *mutationResolver) AddSet(ctx context.Context, exerciseID string, set *model.SetEntryInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set: Invalid Exercise ID")
	}
	exercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set %s", err)
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set: Access Denied")
	}

	dbSet := database.SetEntry{
		ExerciseID: uint(exerciseIDUint),
		Weight:     float32(set.Weight),
		Reps:       uint(set.Reps),
	}
	err = database.AddSet(r.DB, &dbSet)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set")
	}

	return fmt.Sprintf("%d", dbSet.ID), nil
}

// UpdateSet is the resolver for the updateSet field.
func (r *mutationResolver) UpdateSet(ctx context.Context, setID string, set model.UpdateSetEntryInput) (*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.SetEntry{}, err
	}

	var setEntry database.SetEntry
	err = database.GetSet(r.DB, &setEntry, setID)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	exercise := database.Exercise{
		Model: gorm.Model{
			ID: setEntry.ExerciseID,
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set: Access Denied")
	}

	// check optional inputs
	var reps uint
	if set.Reps != nil {
		reps = uint(*set.Reps)
	}
	var weight float32
	if set.Weight != nil {
		weight = float32(*set.Weight)
	}

	updatedSet := database.SetEntry{
		Reps:   reps,
		Weight: weight,
	}
	err = database.UpdateSet(r.DB, setID, &updatedSet)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	return &model.SetEntry{
		ID:     fmt.Sprintf("%d", updatedSet.ID),
		Weight: float64(updatedSet.Weight),
		Reps:   int(updatedSet.Reps),
	}, nil
}

// DeleteSet is the resolver for the deleteSet field.
func (r *mutationResolver) DeleteSet(ctx context.Context, setID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	var setEntry database.SetEntry
	err = database.GetSet(r.DB, &setEntry, setID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	exercise := database.Exercise{
		Model: gorm.Model{
			ID: setEntry.ExerciseID,
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set: Access Denied")
	}

	err = database.DeleteSet(r.DB, setID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	return 1, nil
}

// Sets is the resolver for the sets field.
func (r *queryResolver) Sets(ctx context.Context, exerciseID string) ([]*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.SetEntry{}, err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets: Invalid Exercise ID")
	}
	exercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets: Access Denied")
	}

	var sets []*model.SetEntry
	for _, s := range exercise.Sets {
		sets = append(sets, &model.SetEntry{
			ID:     fmt.Sprintf("%d", s.ID),
			Reps:   int(s.Reps),
			Weight: float64(s.Weight),
		})
	}

	return sets, nil
}


