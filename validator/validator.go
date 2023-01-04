package validator

import (
	"errors"
	"net/mail"

	"github.com/neilZon/workout-logger-api/graph/model"
)

func SignupInputIsValid(s *model.SignupInput) error {
	if _, err := mail.ParseAddress(s.Email); err != nil {
		return errors.New("Not a valid email")
	}

	if len(s.Name) < 2 || len(s.Name) > 50 {
		return errors.New("Name needs to be between 2 and 50 characters")
	}

	if !passwordLongEnough(s.Password) || !hasNumber(s.Password) {
		return errors.New("Password needs at least 1 number and 8 - 32 characters")
	}

	if s.Password != s.ConfirmPassword {
		return errors.New("Passwords don't match")
	}

	return nil
}

func UpdateSetEntryInputIsValid(u *model.UpdateSetEntryInput) error {
	if u.Reps != nil && (*u.Reps > 9999 || *u.Reps < 0) {
		return errors.New("Reps needs to be between 0 and 9999")
	}

	if u.Weight != nil && (*u.Weight > 9999 || *u.Weight < 0) {
		return errors.New("Weight needs to be between 0 and 9999")
	}

	return nil
}

func SetEntryInputIsValid(s *model.SetEntryInput) error {
	if s.Reps < 0 || s.Reps > 9999 {
		return errors.New("Reps needs to be between 0 and 9999")
	}

	if s.Weight < 0 || s.Weight > 9999 {
		return errors.New("Weight needs to be between 0 and 9999")
	}

	return nil
}
