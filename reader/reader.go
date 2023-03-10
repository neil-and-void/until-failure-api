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
		workoutRoutineId := utils.UIntToString(exerciseRoutine.WorkoutRoutineID)
		exerciseRoutineId := utils.UIntToString(exerciseRoutine.ID)
		if _, ok := exerciseRoutinesByWorkoutRoutineId[workoutRoutineId]; ok {
			exerciseRoutinesByWorkoutRoutineId[workoutRoutineId] = append(exerciseRoutinesByWorkoutRoutineId[workoutRoutineId], &model.ExerciseRoutine{
				ID:     exerciseRoutineId,
				Active: exerciseRoutine.Active,
				Name:   exerciseRoutine.Name,
				Sets:   int(exerciseRoutine.Sets),
				Reps:   int(exerciseRoutine.Reps),
			})
		} else {
			exerciseRoutinesByWorkoutRoutineId[workoutRoutineId] = []*model.ExerciseRoutine{
				{
					ID:     exerciseRoutineId,
					Active: exerciseRoutine.Active,
					Name:   exerciseRoutine.Name,
					Sets:   int(exerciseRoutine.Sets),
					Reps:   int(exerciseRoutine.Reps),
				},
			}
		}
	}

	var output []*dataloader.Result
	for _, workoutRoutineKey := range keys {
		if exerciseRoutineSlice, ok := exerciseRoutinesByWorkoutRoutineId[workoutRoutineKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: exerciseRoutineSlice, Error: nil})
		} else {
			output = append(output, &dataloader.Result{Data: []*model.ExerciseRoutine{}, Error: nil})
		}
	}

	return output
}

func (e *ExerciseRoutineReader) GetExerciseRoutines(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	exerciseIds := []string{}
	for _, key := range keys {
		exerciseIds = append(exerciseIds, key.String())
	}

	exercises, _ := database.GetExercisesById(e.DB, exerciseIds)

	// convert to graphql models and store in a dict with exercise id as key
	exerciseRoutineByExerciseId := map[string]*model.ExerciseRoutine{}
	for _, exercise := range *exercises {
		exerciseId := strconv.Itoa(int(exercise.ID))
		exerciseRoutineId := strconv.Itoa(int(exercise.ExerciseRoutineID))

		exerciseRoutineByExerciseId[exerciseId] = &model.ExerciseRoutine{
			ID:     exerciseRoutineId,
			Name:   exercise.ExerciseRoutine.Name,
			Active: exercise.ExerciseRoutine.Active,
			Sets:   int(exercise.ExerciseRoutine.Sets),
			Reps:   int(exercise.ExerciseRoutine.Reps),
		}
	}

	var output []*dataloader.Result
	for _, exerciseRoutineKey := range keys {
		exerciseRoutine, ok := exerciseRoutineByExerciseId[exerciseRoutineKey.String()]
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
		workoutSessionId := utils.UIntToString(exercise.WorkoutSessionID)
		exerciseId := utils.UIntToString(exercise.ID)
		if _, ok := exerciseSlicesByWorkoutSession[workoutSessionId]; ok {
			exerciseSlicesByWorkoutSession[workoutSessionId] = append(exerciseSlicesByWorkoutSession[workoutSessionId], &model.Exercise{
				ID:    exerciseId,
				Notes: exercise.Notes,
			})
		} else {
			exerciseSlicesByWorkoutSession[workoutSessionId] = []*model.Exercise{
				{
					ID:    exerciseId,
					Notes: exercise.Notes,
				},
			}
		}
	}

	var output []*dataloader.Result
	for _, workoutSessionKey := range keys {
		if exerciseRoutineSlice, ok := exerciseSlicesByWorkoutSession[workoutSessionKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: exerciseRoutineSlice, Error: nil})
		} else {
			output = append(output, &dataloader.Result{Data: []*model.Exercise{}, Error: nil})
		}
	}

	return output
}

func (s *SetEntrySliceReader) GetSetEntrySlices(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	exerciseIds := []string{}
	for _, key := range keys {
		exerciseIds = append(exerciseIds, key.String())
	}

	setEntries, _ := database.GetSetsByExerciseId(s.DB, exerciseIds)
	setEntrySlicesByExerciseId := map[string][]*model.SetEntry{}
	for _, setEntry := range *setEntries {
		exerciseId := utils.UIntToString(setEntry.ExerciseID)
		setEntryId := utils.UIntToString(setEntry.ID)
		if _, ok := setEntrySlicesByExerciseId[exerciseId]; ok {
			setEntrySlicesByExerciseId[exerciseId] = append(setEntrySlicesByExerciseId[exerciseId], &model.SetEntry{
				ID:     setEntryId,
				Weight: float64(setEntry.Weight),
				Reps:   int(setEntry.Reps),
			})
		} else {
			setEntrySlicesByExerciseId[exerciseId] = []*model.SetEntry{
				{
					ID:     setEntryId,
					Weight: float64(setEntry.Weight),
					Reps:   int(setEntry.Reps),
				},
			}
		}
	}

	var output []*dataloader.Result
	for _, exerciseKey := range keys {
		if setEntrySlice, ok := setEntrySlicesByExerciseId[exerciseKey.String()]; ok {
			output = append(output, &dataloader.Result{Data: setEntrySlice, Error: nil})
		} else {
			output = append(output, &dataloader.Result{Data: []*model.SetEntry{}, Error: nil})
		}
	}

	return output
}
