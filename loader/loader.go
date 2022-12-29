package loader

import (
	"github.com/graph-gophers/dataloader"
)

type Loaders struct {
	WorkoutRoutineLoader  *dataloader.Loader
	ExerciseRoutineLoader *dataloader.Loader
	SetEntryLoader        *dataloader.Loader
}
