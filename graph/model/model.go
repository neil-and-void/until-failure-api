package model

import "time"

type WorkoutRoutine struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Active           bool               `json:"active"`
	ExerciseRoutines []*ExerciseRoutine `json:"exerciseRoutines"`
}

type WorkoutSession struct {
	ID             string         `json:"id"`
	Start          time.Time      `json:"start"`
	End            *time.Time     `json:"end"`
	WorkoutRoutine WorkoutRoutine `json:"workoutRoutine"`
	Exercises      []*Exercise    `json:"exercises"`
}

type Exercise struct {
	ID              string          `json:"id"`
	ExerciseRoutine ExerciseRoutine `json:"exerciseRoutine"`
	Prev            *PrevExercise   `json:"prev"`
	Sets            []*SetEntry     `json:"sets"`
	Notes           string          `json:"notes"`
}

type PrevExercise struct {
	ID    string      `json:"id"`
	Sets  []*SetEntry `json:"sets"`
	Notes string      `json:"notes"`
}
