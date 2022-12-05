// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type AuthResult interface {
	IsAuthResult()
}

type AuthError struct {
	Message string `json:"message"`
}

func (AuthError) IsAuthResult() {}

type AuthSuccess struct {
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
}

func (AuthSuccess) IsAuthResult() {}

type Exercise struct {
	ID                string      `json:"id"`
	ExerciseRoutineID string      `json:"exerciseRoutineId"`
	Sets              []*SetEntry `json:"sets"`
	Notes             string      `json:"notes"`
}

type ExerciseInput struct {
	ExerciseRoutineID string           `json:"exerciseRoutineId"`
	Notes             string           `json:"notes"`
	SetEntries        []*SetEntryInput `json:"setEntries"`
}

type ExerciseRoutine struct {
	ID     string `json:"id"`
	Active bool   `json:"active"`
	Name   string `json:"name"`
	Sets   int    `json:"sets"`
	Reps   int    `json:"reps"`
}

type ExerciseRoutineInput struct {
	Name string `json:"name"`
	Sets int    `json:"sets"`
	Reps int    `json:"reps"`
}

type RefreshSuccess struct {
	AccessToken string `json:"accessToken"`
}

type SetEntry struct {
	ID     string  `json:"id"`
	Weight float64 `json:"weight"`
	Reps   int     `json:"reps"`
}

type SetEntryInput struct {
	Weight float64 `json:"weight"`
	Reps   int     `json:"reps"`
}

type UpdateExerciseInput struct {
	Notes string `json:"notes"`
}

type UpdateExerciseRoutineInput struct {
	ID   *string `json:"id"`
	Name string  `json:"name"`
	Sets int     `json:"sets"`
	Reps int     `json:"reps"`
}

type UpdateSetEntryInput struct {
	Weight *float64 `json:"weight"`
	Reps   *int     `json:"reps"`
}

type UpdateWorkoutRoutineInput struct {
	ID               string                        `json:"id"`
	Name             string                        `json:"name"`
	ExerciseRoutines []*UpdateExerciseRoutineInput `json:"exerciseRoutines"`
}

type UpdateWorkoutSessionInput struct {
	Start *time.Time `json:"start"`
	End   *time.Time `json:"end"`
}

type UpdatedExercise struct {
	ID    string `json:"id"`
	Notes string `json:"notes"`
}

type UpdatedWorkoutSession struct {
	ID    string     `json:"id"`
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type WorkoutRoutine struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Active           bool               `json:"active"`
	ExerciseRoutines []*ExerciseRoutine `json:"exerciseRoutines"`
}

type WorkoutRoutineInput struct {
	Name             string                  `json:"name"`
	ExerciseRoutines []*ExerciseRoutineInput `json:"exerciseRoutines"`
}

type WorkoutSession struct {
	ID               string      `json:"id"`
	Start            time.Time   `json:"start"`
	End              *time.Time  `json:"end"`
	WorkoutRoutineID string      `json:"workoutRoutineId"`
	Exercises        []*Exercise `json:"exercises"`
}

type WorkoutSessionInput struct {
	WorkoutRoutineID string           `json:"workoutRoutineId"`
	Start            time.Time        `json:"start"`
	End              *time.Time       `json:"end"`
	Exercises        []*ExerciseInput `json:"exercises"`
}
