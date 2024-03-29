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
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// AddExerciseRoutine is the resolver for the addExerciseRoutine field.
func (r *mutationResolver) AddExerciseRoutine(ctx context.Context, workoutRoutineID string, exerciseRoutine model.ExerciseRoutineInput) (*model.ExerciseRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.ExerciseRoutine{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.ExerciseRoutine{}, err
	}

	if exerciseRoutine.Sets > 20 {
		return &model.ExerciseRoutine{}, gqlerror.Errorf("Exercise routine cannot have more than 20 sets")
	}

	if exerciseRoutine.Sets < 0 {
		return &model.ExerciseRoutine{}, gqlerror.Errorf("sets cannot be a negative number")
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return &model.ExerciseRoutine{}, gqlerror.Errorf("Error Adding Exercise Routine: Access Denied")
	}

	workoutRoutineIDUint, err := strconv.ParseUint(workoutRoutineID, 10, strconv.IntSize)
	if err != nil {
		return &model.ExerciseRoutine{}, gqlerror.Errorf("Error Adding Exercise Routine")
	}
	dbExerciseRoutine := &database.ExerciseRoutine{
		Name:             exerciseRoutine.Name,
		Sets:             uint(exerciseRoutine.Sets),
		Reps:             uint(exerciseRoutine.Reps),
		WorkoutRoutineID: uint(workoutRoutineIDUint),
	}
	err = database.AddExerciseRoutine(r.DB, dbExerciseRoutine)
	if err != nil {
		return &model.ExerciseRoutine{}, gqlerror.Errorf("Error Adding Exercise Routine")
	}

	loaders := middleware.GetLoaders(ctx)
	loaders.ExerciseRoutineSliceLoader.Clear(ctx, dataloader.StringKey(workoutRoutineID))

	return &model.ExerciseRoutine{
		ID:     utils.UIntToString(dbExerciseRoutine.ID),
		Active: dbExerciseRoutine.Active,
		Name:   dbExerciseRoutine.Name,
		Reps:   int(dbExerciseRoutine.Reps),
		Sets:   int(dbExerciseRoutine.Sets),
	}, nil
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *queryResolver) ExerciseRoutines(ctx context.Context, workoutRoutineID string) ([]*model.ExerciseRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.ExerciseRoutine{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return []*model.ExerciseRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine: Access Denied")
	}

	dbExerciseRoutines, err := database.GetExerciseRoutines(r.DB, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine")
	}

	exerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range *dbExerciseRoutines {
		exerciseRoutines = append(exerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Sets: int(er.Sets),
			Reps: int(er.Reps),
		})
	}

	return exerciseRoutines, nil
}

// DeleteExerciseRoutine is the resolver for the deleteExerciseRoutine field.
func (r *mutationResolver) DeleteExerciseRoutine(ctx context.Context, exerciseRoutineID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return 0, err
	}

	exerciseRoutine := database.ExerciseRoutine{}
	err = database.GetExerciseRoutine(r.DB, exerciseRoutineID, &exerciseRoutine)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine")
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, fmt.Sprintf("%d", exerciseRoutine.WorkoutRoutineID))
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine: Access Denied")
	}

	err = database.DeleteExerciseRoutine(r.DB, exerciseRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine")
	}

	return 1, nil
}

// ExerciseRoutine is the resolver for the exerciseRoutine field.
func (r *exerciseResolver) ExerciseRoutine(ctx context.Context, obj *model.Exercise) (*model.ExerciseRoutine, error) {
	loaders := middleware.GetLoaders(ctx)
	thunk := loaders.ExerciseRoutineLoader.Load(ctx, dataloader.StringKey(obj.ID))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.ExerciseRoutine), nil
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *workoutRoutineResolver) ExerciseRoutines(ctx context.Context, obj *model.WorkoutRoutine) ([]*model.ExerciseRoutine, error) {
	loaders := middleware.GetLoaders(ctx)
	thunk := loaders.ExerciseRoutineSliceLoader.Load(ctx, dataloader.StringKey(obj.ID))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.([]*model.ExerciseRoutine), nil
}
