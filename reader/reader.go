// Package defines readers for dataloaders to
// use to read data in batches

package reader

import (
	"context"
	"fmt"
	"strconv"

	"github.com/graph-gophers/dataloader"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"gorm.io/gorm"
)

type WorkoutRoutineReader struct {
	DB *gorm.DB
}

type ExerciseRoutineReader struct {
	DB *gorm.DB
}

type SetEntryReader struct {
	DB *gorm.DB
}

func (w *WorkoutRoutineReader) GetWorkoutRoutines(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	workoutSessionIds := []string{}
	for _, key := range keys {
		workoutSessionIds = append(workoutSessionIds, key.String())
	}

	workoutSessions, _ := database.GetWorkoutSessionsById(w.DB, workoutSessionIds)
	workoutRoutineById := map[string]*model.WorkoutRoutine{}
	for _, workoutSession := range *workoutSessions {
		workoutSessionId := strconv.Itoa(int(workoutSession.ID))
		workoutRoutineId := strconv.Itoa(int(workoutSession.WorkoutRoutine.ID))
		workoutRoutineById[workoutSessionId] = &model.WorkoutRoutine{
			ID:     workoutRoutineId,
			Name:   workoutSession.WorkoutRoutine.Name,
			Active: workoutSession.WorkoutRoutine.Active,
		}
	}

	var output []*dataloader.Result
	for _, workoutSessionKey := range keys {
		workoutRoutine, ok := workoutRoutineById[workoutSessionKey.String()]
		if ok {
			output = append(output, &dataloader.Result{Data: workoutRoutine, Error: nil})
		} else {
			err := fmt.Errorf("workout routine not found %s", workoutSessionKey.String())
			output = append(output, &dataloader.Result{Data: nil, Error: err})
		}
	}
	return output
}

func (e *ExerciseRoutineReader) GetExerciseRoutines(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	exerciseRoutineIds := []string{}
	for _, key := range keys {
		exerciseRoutineIds = append(exerciseRoutineIds, key.String())
	}
	exercises, _ := database.GetExercisesById(e.DB, exerciseRoutineIds)

	// convert to graphql models and store in a dict with workout routine id as key
	exerciseRoutineById := map[string]*model.ExerciseRoutine{}
	for _, exercise := range *exercises {
		id := strconv.Itoa(int(exercise.ID))
		exerciseRoutineById[id] = &model.ExerciseRoutine{
			ID:     id,
			Name:   exercise.ExerciseRoutine.Name,
			Active: exercise.ExerciseRoutine.Active,
			Sets:   int(exercise.ExerciseRoutine.Sets),
			Reps:   int(exercise.ExerciseRoutine.Reps),
		}
	}

	var output []*dataloader.Result
	for _, exerciseRoutineKey := range keys {
		exerciseRoutine, ok := exerciseRoutineById[exerciseRoutineKey.String()]
		if ok {
			output = append(output, &dataloader.Result{Data: exerciseRoutine, Error: nil})
		} else {
			err := fmt.Errorf("exercise routine not found %s", exerciseRoutineKey.String())
			output = append(output, &dataloader.Result{Data: nil, Error: err})
		}
	}
	
	return output
}

func (e *SetEntryReader) GetSetEntries(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	panic("something")
}
