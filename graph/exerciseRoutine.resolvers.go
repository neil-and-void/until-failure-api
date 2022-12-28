package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// AddExerciseRoutine is the resolver for the addExerciseRoutine field.
func (r *mutationResolver) AddExerciseRoutine(ctx context.Context, workoutRoutineID string, exerciseRoutine model.ExerciseRoutineInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine: Access Denied")
	}

	workoutRoutineIDUint, err := strconv.ParseUint(workoutRoutineID, 10, strconv.IntSize)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine")
	}
	dbExerciseRoutine := &database.ExerciseRoutine{
		Name:             exerciseRoutine.Name,
		Sets:             uint(exerciseRoutine.Sets),
		Reps:             uint(exerciseRoutine.Reps),
		WorkoutRoutineID: uint(workoutRoutineIDUint),
	}
	err = database.AddExerciseRoutine(r.DB, dbExerciseRoutine)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine")
	}

	return fmt.Sprintf("%d", dbExerciseRoutine.ID), nil
}

// ExerciseRoutine is the resolver for the exerciseRoutine field.
func (r *exerciseResolver) ExerciseRoutine(ctx context.Context, obj *model.Exercise) (*model.ExerciseRoutine, error) {
	panic(fmt.Errorf("not implemented: ExerciseRoutine - exerciseRoutine"))
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *queryResolver) ExerciseRoutines(ctx context.Context, workoutRoutineID string) ([]*model.ExerciseRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.ExerciseRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine: Access Denied")
	}

	erdb, err := database.GetExerciseRoutines(r.DB, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine")
	}

	exerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range erdb {
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

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *workoutRoutineResolver) ExerciseRoutines(ctx context.Context, obj *model.WorkoutRoutine) ([]*model.ExerciseRoutine, error) {
	fmt.Println("THIS THING WAS CALLED", *obj)
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.ExerciseRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, obj.ID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine: Access Denied")
	}

	erdb, err := database.GetExerciseRoutines(r.DB, obj.ID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine")
	}

	exerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range erdb {
		exerciseRoutines = append(exerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Sets: int(er.Sets),
			Reps: int(er.Reps),
		})
	}

	return exerciseRoutines, nil
}

