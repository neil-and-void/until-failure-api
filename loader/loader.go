package loader

import (
	"github.com/graph-gophers/dataloader"
)

// Struct of batch loaders to reduce db calls
type Loaders struct {
	WorkoutRoutineLoader  *dataloader.Loader
	ExerciseRoutineLoader *dataloader.Loader
	ExerciseRoutineSliceLoader *dataloader.Loader
	ExerciseSliceLoader *dataloader.Loader
	SetEntrySliceLoader        *dataloader.Loader
}
