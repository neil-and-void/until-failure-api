package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/graph-gophers/dataloader"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/errors"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/neilZon/workout-logger-api/validator"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
)

// CreateWorkoutRoutine is the resolver for the createWorkoutRoutine field.
func (r *mutationResolver) CreateWorkoutRoutine(ctx context.Context, routine model.WorkoutRoutineInput) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	// validate input
	if len([]rune(routine.Name)) <= 2 {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Invalid Routine Name Length")
	}

	if len(routine.ExerciseRoutines) > 20 {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("workout routine can only have 20 exercise routines max")
	}

	for _, exerciseRoutine := range routine.ExerciseRoutines {
		validator.ExerciseRoutineIsValid(&model.ExerciseRoutine{
			ID:   "", // blank string to pass to validator
			Name: exerciseRoutine.Name,
			Reps: exerciseRoutine.Reps,
			Sets: exerciseRoutine.Sets,
		})
	}

	exerciseRoutines := make([]database.ExerciseRoutine, 0)
	for _, er := range routine.ExerciseRoutines {
		exerciseRoutines = append(exerciseRoutines, database.ExerciseRoutine{Name: er.Name, Reps: uint(er.Reps), Sets: uint(er.Sets)})
	}

	wr := &database.WorkoutRoutine{
		Name:             routine.Name,
		ExerciseRoutines: exerciseRoutines,
		UserID:           u.ID,
	}

	res := database.CreateWorkoutRoutine(r.DB, wr)
	if res.Error != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Creating Workout Routine")
	}

	dbExerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range wr.ExerciseRoutines {
		dbExerciseRoutines = append(dbExerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Sets: int(er.Sets),
			Reps: int(er.Reps),
		})
	}

	return &model.WorkoutRoutine{
		ID:               fmt.Sprintf("%d", wr.ID),
		Name:             wr.Name,
		ExerciseRoutines: []*model.ExerciseRoutine{},
		Active:           wr.Active,
	}, nil
}

// WorkoutRoutines is the resolver for the workoutRoutines field.
func (r *queryResolver) WorkoutRoutines(ctx context.Context, limit int, after *string) (*model.WorkoutRoutineConnection, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutineConnection{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutRoutineConnection{}, err
	}

	if limit <= 0 || limit > 50 {
		return &model.WorkoutRoutineConnection{}, fmt.Errorf(errors.GetWorkoutRoutinesError, "limit needs to be between 1 to 50")
	}

	var dbWorkoutRoutines []database.WorkoutRoutine
	cursor := ""
	if after != nil && *after != "" {
		cursor = *after
	}

	dbWorkoutRoutines, err = database.GetWorkoutRoutines(r.DB, utils.UIntToString(u.ID), cursor, limit)

	if err != nil {
		return &model.WorkoutRoutineConnection{}, gqlerror.Errorf("Error Getting Workout Routine")
	}

	var edges []*model.WorkoutRoutineEdge
	for _, workoutRoutine := range dbWorkoutRoutines {
		edges = append(edges, &model.WorkoutRoutineEdge{
			Cursor: utils.UIntToString(workoutRoutine.ID),
			Node: &model.WorkoutRoutine{
				ID:     utils.UIntToString(workoutRoutine.ID),
				Name:   workoutRoutine.Name,
				Active: workoutRoutine.Active,
			},
		})
	}

	return &model.WorkoutRoutineConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage: true,
		},
	}, nil
}

// WorkoutRoutine is the resolver for the workoutRoutine field.
func (r *queryResolver) WorkoutRoutine(ctx context.Context, workoutRoutineID string) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Getting Workout Routine: Access Denied")
	}

	workoutRoutine, err := database.GetWorkoutRoutine(r.DB, workoutRoutineID)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Getting Workout Routine")
	}

	return &model.WorkoutRoutine{
		ID:     fmt.Sprintf("%d", workoutRoutine.ID),
		Name:   workoutRoutine.Name,
		Active: workoutRoutine.Active,
	}, nil
}

// UpdateWorkoutRoutine is the resolver for the updateWorkoutRoutine field.
func (r *mutationResolver) UpdateWorkoutRoutine(ctx context.Context, workoutRoutine model.UpdateWorkoutRoutineInput) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	for _, exerciseRoutine := range workoutRoutine.ExerciseRoutines {
		err = validator.ExerciseRoutineIsValid(&model.ExerciseRoutine{
			ID:   "", // blank string to pass to validator
			Name: exerciseRoutine.Name,
			Reps: exerciseRoutine.Reps,
			Sets: exerciseRoutine.Sets,
		})

		if err != nil {
			return &model.WorkoutRoutine{}, err
		}
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutine.ID)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Updating Workout Routine: Access Denied")
	}

	var exerciseRoutines []*database.ExerciseRoutine
	for _, er := range workoutRoutine.ExerciseRoutines {
		// newly added exercises won't have an ID
		// nil ID indicates that this exercise should be created, otherwise update
		// the exercise that has that ID
		var model gorm.Model
		if er.ID != nil {
			num, err := strconv.ParseUint(*er.ID, 10, strconv.IntSize)
			if err != nil {
				panic(err)
			}
			model.ID = uint(num)
		}

		workoutRoutineIDUint, err := strconv.ParseUint(workoutRoutine.ID, 10, strconv.IntSize)
		if err != nil {
			panic(err)
		}

		exerciseRoutines = append(exerciseRoutines, &database.ExerciseRoutine{
			Model:            model,
			Name:             er.Name,
			Sets:             uint(er.Sets),
			Reps:             uint(er.Reps),
			WorkoutRoutineID: uint(workoutRoutineIDUint),
		})
	}

	err = database.UpdateWorkoutRoutine(r.DB, workoutRoutine.ID, workoutRoutine.Name, exerciseRoutines)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Updating Workout Routine")
	}

	// invalidate cache to return freshly updated exercise routines
	loaders := middleware.GetLoaders(ctx)
	loaders.ExerciseRoutineSliceLoader.Clear(ctx, dataloader.StringKey(workoutRoutine.ID))

	return &model.WorkoutRoutine{
		ID:   workoutRoutine.ID,
		Name: workoutRoutine.Name,
	}, nil
}

// DeleteWorkoutRoutine is the resolver for the deleteWorkoutRoutine field.
func (r *mutationResolver) DeleteWorkoutRoutine(ctx context.Context, workoutRoutineID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	err = middleware.VerifyUser(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return 0, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Routine: Access Denied")
	}

	err = database.DeleteWorkoutRoutine(r.DB, workoutRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Routine")
	}

	return 1, nil
}

// WorkoutRoutine is the resolver for the workoutRoutine field.
func (r *workoutSessionResolver) WorkoutRoutine(ctx context.Context, obj *model.WorkoutSession) (*model.WorkoutRoutine, error) {
	loaders := middleware.GetLoaders(ctx)
	thunk := loaders.WorkoutRoutineLoader.Load(ctx, dataloader.StringKey(obj.ID))
	result, err := thunk()
	if err != nil {
		return nil, err
	}
	return result.(*model.WorkoutRoutine), nil
}
