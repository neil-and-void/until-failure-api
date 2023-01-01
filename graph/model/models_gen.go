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

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
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

type SignupInput struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
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

type WorkoutRoutineConnection struct {
	Edges    []*WorkoutRoutineEdge `json:"edges"`
	PageInfo *PageInfo             `json:"pageInfo"`
}

type WorkoutRoutineEdge struct {
	Node   *WorkoutRoutine `json:"node"`
	Cursor string          `json:"cursor"`
}

type WorkoutRoutineInput struct {
	Name             string                  `json:"name"`
	ExerciseRoutines []*ExerciseRoutineInput `json:"exerciseRoutines"`
}

type WorkoutSessionConnection struct {
	Edges    []*WorkoutSessionEdge `json:"edges"`
	PageInfo *PageInfo             `json:"pageInfo"`
}

type WorkoutSessionEdge struct {
	Node   *WorkoutSession `json:"node"`
	Cursor string          `json:"cursor"`
}

type WorkoutSessionInput struct {
	WorkoutRoutineID string           `json:"workoutRoutineId"`
	Start            time.Time        `json:"start"`
	End              *time.Time       `json:"end"`
	Exercises        []*ExerciseInput `json:"exercises"`
}
