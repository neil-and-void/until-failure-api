package validator

import (
	"errors"
	"net/mail"

	"github.com/neilZon/workout-logger-api/graph/model"
)

func SignupInputIsValid(s *model.SignupInput) error {
	if _, err := mail.ParseAddress(s.Email); err != nil {
		return errors.New("not a valid email")
	}

	if len(s.Name) < 2 || len(s.Name) > 50 {
		return errors.New("name needs to be between 2 and 50 characters")
	}

	if !passwordLongEnough(s.Password) || !hasNumber(s.Password) {
		return errors.New("password needs at least 1 number and 8 - 32 characters")
	}

	if s.Password != s.ConfirmPassword {
		return errors.New("passwords don't match")
	}

	return nil
}

func ValidateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("not a valid email")
	}
	return nil
}

func UpdateSetEntryInputIsValid(u *model.UpdateSetEntryInput) error {
	if u.Reps != nil && (*u.Reps > 9999 || *u.Reps < 0) {
		return errors.New("reps needs to be between 0 and 9999")
	}

	if u.Weight != nil && (*u.Weight > 9999 || *u.Weight < 0) {
		return errors.New("weight needs to be between 0 and 9999")
	}

	return nil
}

func SetEntryInputIsValid(s *model.SetEntry) error {
	if s.Reps < 0 || s.Reps > 9999 {
		return errors.New("reps needs to be between 0 and 9999")
	}

	if s.Weight < 0 || s.Weight > 9999 {
		return errors.New("weight needs to be between 0 and 9999")
	}

	return nil
}

func ExerciseIsVaid(exercise *model.Exercise) error {
	if len(exercise.Sets) > 20 {
		return errors.New("exercise cannot have more than 20 sets")
	}

	for _, set := range exercise.Sets {
		if err := SetEntryInputIsValid(set); err != nil {
			return err
		}
	}

	if len(exercise.Notes) > 512 {
		return errors.New("max length of notes is 512 character")
	}

	return nil
}

func ExerciseRoutineIsValid(exerciseRoutine *model.ExerciseRoutine) error { return nil }

func WorkoutSessionIsValid(workoutSession *model.WorkoutSession) error { return nil }

func WorkoutRoutineIsValid(workoutRoutine *model.WorkoutRoutine) error { return nil }
