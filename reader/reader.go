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
	"github.com/neilZon/workout-logger-api/utils"
	"gorm.io/gorm"
)

type WorkoutRoutineReader struct {
	DB *gorm.DB
}

type ExerciseRoutineSliceReader struct {
	DB *gorm.DB
}

type ExerciseRoutineReader struct {
	DB *gorm.DB
}

type ExerciseSliceReader struct {
	DB *gorm.DB
}

type PrevExerciseSliceReader struct {
	DB *gorm.DB
}

type SetEntrySliceReader struct {
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

func (e *ExerciseRoutineSliceReader) GetExerciseRoutineSlices(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	workoutRoutineIds := []string{}
	for _, key := range keys {
		workoutRoutineIds = append(workoutRoutineIds, key.String())
	}
	exerciseRoutines, _ := database.GetExerciseRoutinesByWorkoutRoutineId(e.DB, workoutRoutineIds)
	exerciseRoutinesByWorkoutRoutineId := map[string][]*model.ExerciseRoutine{}
	for _, exerciseRoutine := range *exerciseRoutines {
		id := utils.UIntToString(exerciseRoutine.WorkoutRoutineID)
		if _, ok := exerciseRoutinesByWorkoutRoutineId[id]; ok {
			exerciseRoutinesByWorkoutRoutineId[id] = append(exerciseRoutinesByWorkoutRoutineId[id], &model.ExerciseRoutine{
				ID: id,
				Active: exerciseRoutine.Active,
				Name: exerciseRoutine.Name,
				Sets: int(exerciseRoutine.Sets),
				Reps: int(exerciseRoutine.Reps),
			})
		} else {
			exerciseRoutinesByWorkoutRoutineId[id] = []*model.ExerciseRoutine{
				{
					ID: id,
					Active: exerciseRoutine.Active,
					Name: exerciseRoutine.Name,
					Sets: int(exerciseRoutine.Sets),
					Reps: int(exerciseRoutine.Reps),
				},
			}
		}
	}

	var output []*dataloader.Result
	for _, workoutRoutineKey := range keys {
		if exerciseRoutineSlice, ok := exerciseRoutinesByWorkoutRoutineId[workoutRoutineKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: exerciseRoutineSlice, Error: nil})
		} else {
			err := fmt.Errorf("exercise routine slice not found %s", workoutRoutineKey.String())
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

func (e *ExerciseSliceReader) GetExerciseSlices(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	workoutSessionIds := []string{}
	for _, key := range keys {
		workoutSessionIds = append(workoutSessionIds, key.String())
	}

	exercises, _ := database.GetExercisesByWorkoutSessionId(e.DB, workoutSessionIds)
	exerciseSlicesByWorkoutSession := map[string][]*model.Exercise{}
	for _, exercise := range *exercises {
		id := fmt.Sprintf("%d", exercise.WorkoutSessionID)

		if _, ok := exerciseSlicesByWorkoutSession[id]; ok {
			exerciseSlicesByWorkoutSession[id] = append(exerciseSlicesByWorkoutSession[id], &model.Exercise{
				ID: id,
				Notes: exercise.Notes,
			})
		}
	}

	var output []*dataloader.Result
	for _, workoutSessionKey := range keys {
		if exerciseRoutineSlice, ok := exerciseSlicesByWorkoutSession[workoutSessionKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: exerciseRoutineSlice, Error: nil})
		} else {
			err := fmt.Errorf("exercise slice not found %s", workoutSessionKey.String())
			output = append(output, &dataloader.Result{Data: nil, Error: err})	
		}	
	}

	return output
}

func (p *PrevExerciseSliceReader) GetPrevExerciseSlices(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	panic("unimplemented")
}

func (s *SetEntrySliceReader) GetSetEntrySlices(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {	
	exerciseIds := []string{}
	for _, key := range keys {
		exerciseIds = append(exerciseIds, key.String())
	}

	setEntries, _ := database.GetSetsByExerciseId(s.DB, exerciseIds)
	exerciseSlicesByWorkoutSession := map[string][]*model.SetEntry{}
	for _, setEntry := range *setEntries {
		id := fmt.Sprintf("%d", setEntry.ExerciseID)

		if _, ok := exerciseSlicesByWorkoutSession[id]; ok {
			exerciseSlicesByWorkoutSession[id] = append(exerciseSlicesByWorkoutSession[id], &model.SetEntry{
				ID: id,
				Weight: float64(setEntry.Weight),
				Reps: int(setEntry.Reps),
			})
		}
	}

	var output []*dataloader.Result
	for _, exerciseKey := range keys {
		if setEntrySlice, ok := exerciseSlicesByWorkoutSession[exerciseKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: setEntrySlice, Error: nil})
		} else {
			err := fmt.Errorf("exercise slice not found %s", exerciseKey.String())
			output = append(output, &dataloader.Result{Data: nil, Error: err})	
		}		
	}

	return output	
}
