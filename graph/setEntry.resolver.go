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

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.SetEntry{}, err
	}

	if err := validator.SetEntryInputIsValid(&model.SetEntry{Weight: set.Weight, Reps: set.Reps}); err != nil {
		return &model.SetEntry{}, err
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

	var weight *float32
	if set.Weight != nil {
		w := float32(*set.Weight)
		weight = &w
	}
	var reps *uint
	if set.Weight != nil {
		r := uint(*set.Reps)
		reps = &r
	}

	dbSet := database.SetEntry{
		ExerciseID: uint(exerciseIDUint),
		Weight:     weight,
		Reps:       reps,
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
		Weight: set.Weight,
		Reps:   set.Reps,
	}, nil
}

// Sets is the resolver for the sets field.
func (r *queryResolver) Sets(ctx context.Context, exerciseID string) ([]*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.SetEntry{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
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

		var weight *float64
		if s.Weight != nil {
			w := float64(*s.Weight)
			weight = &w
		}
		var reps *int
		if s.Weight != nil {
			r := int(*s.Reps)
			reps = &r
		}

		sets = append(sets, &model.SetEntry{
			ID:     fmt.Sprintf("%d", s.ID),
			Reps:   reps,
			Weight: weight,
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

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.SetEntry{}, err
	}

	if set.Reps != nil && (*set.Reps < 0 || *set.Reps > 9999) {
		return &model.SetEntry{}, gqlerror.Errorf("Reps needs to be between 0 and 9999")
	}

	if set.Weight != nil && (*set.Weight < 0 || *set.Weight > 9999) {
		return &model.SetEntry{}, gqlerror.Errorf("Weight needs to be between 0 and 9999")
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
	var dbReps uint
	if set.Reps != nil {
		dbReps = uint(*set.Reps)
	}
	var dbWeight float32
	if set.Weight != nil {
		dbWeight = float32(*set.Weight)
	}

	updatedSet := database.SetEntry{
		Reps:   &dbReps,
		Weight: &dbWeight,
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
		Weight: set.Weight,
		Reps:   set.Reps,
	}, nil
}

// DeleteSet is the resolver for the deleteSet field.
func (r *mutationResolver) DeleteSet(ctx context.Context, setID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
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
