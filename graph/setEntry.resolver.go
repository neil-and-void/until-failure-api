package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/graph-gophers/dataloader"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/neilZon/workout-logger-api/validator"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

// AddSet is the resolver for the addSet field.
func (r *mutationResolver) AddSet(ctx context.Context, exerciseID string, set model.SetEntryInput) (*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.SetEntry{}, err
	}

	if set.Reps < 0 || 999 < set.Reps {
		return &model.SetEntry{}, gqlerror.Errorf("reps needs to be between 0 and 999")
	}

	if set.Weight < 0 || 9999 < set.Weight {

	}

	if err := validator.SetEntryInputIsValid(&set); err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("weight needs to be between 0 and 9999")
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Adding Set: Invalid Exercise ID")
	}
	exercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &exercise, false)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Adding Set: %s", err)
	}
	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Adding Set: Access Denied")
	}

	dbSet := database.SetEntry{
		ExerciseID: uint(exerciseIDUint),
		Weight:     float32(set.Weight),
		Reps:       uint(set.Reps),
	}
	err = database.AddSet(r.DB, &dbSet)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Adding Set")
	}

	// invalidate set entry resolver dataloader cache
	loaders := middleware.GetLoaders(ctx)
	loaders.SetEntrySliceLoader.Clear(ctx, dataloader.StringKey(exerciseID))

	return &model.SetEntry{
		ID:     utils.UIntToString(dbSet.ID),
		Weight: float64(dbSet.Weight),
		Reps:   int(dbSet.Reps),
	}, nil
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
	err = database.GetExercise(r.DB, &exercise, true)
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

// UpdateSet is the resolver for the updateSet field.
func (r *mutationResolver) UpdateSet(ctx context.Context, setID string, set model.UpdateSetEntryInput) (*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.SetEntry{}, err
	}

	if err := validator.UpdateSetEntryInputIsValid(&set); err != nil {
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
	err = database.GetExercise(r.DB, &exercise, false)
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

	// invalidate set entry resolver dataloader cache
	loaders := middleware.GetLoaders(ctx)
	loaders.SetEntrySliceLoader.Clear(ctx, dataloader.StringKey(fmt.Sprintf("%d", exercise.ID)))

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
	err = database.GetExercise(r.DB, &exercise, false)
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

	// invalidate set entry resolver dataloader cache
	loaders := middleware.GetLoaders(ctx)
	loaders.SetEntrySliceLoader.Clear(ctx, dataloader.StringKey(fmt.Sprintf("%d", exercise.ID)))

	return 1, nil
}

// Sets is the resolver for the sets field.
func (r *exerciseResolver) Sets(ctx context.Context, obj *model.Exercise) ([]*model.SetEntry, error) {
	loaders := middleware.GetLoaders(ctx)
	thunk := loaders.SetEntrySliceLoader.Load(ctx, dataloader.StringKey(obj.ID))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]*model.SetEntry), nil
}
