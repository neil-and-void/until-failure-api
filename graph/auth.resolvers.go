package graph

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/neilZon/workout-logger-api/config"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/mail"
	"github.com/neilZon/workout-logger-api/token"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/neilZon/workout-logger-api/validator"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, loginInput model.LoginInput) (*model.AuthResult, error) {
	dbUser, err := database.GetUserByEmail(r.DB, loginInput.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.AuthResult{}, gqlerror.Errorf("Email does not exist")
	}
	if err != nil {
		return &model.AuthResult{}, gqlerror.Errorf("Error Logging In")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginInput.Password)); err != nil {
		return &model.AuthResult{}, gqlerror.Errorf("Incorrect Password")
	}
	c := &token.Credentials{
		ID:    dbUser.ID,
		Email: dbUser.Email,
		Name:  dbUser.Name,
	}

	refreshToken := token.Sign(c, []byte(os.Getenv(config.REFRESH_SECRET)), config.REFRESH_TTL)
	accessToken := token.Sign(c, []byte(os.Getenv(config.ACCESS_SECRET)), config.ACCESS_TTL)

	return &model.AuthResult{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

// Signup is the resolver for the signup field.
func (r *mutationResolver) Signup(ctx context.Context, signupInput model.SignupInput) (*model.AuthResult, error) {
	if err := validator.SignupInputIsValid(&signupInput); err != nil {
		return &model.AuthResult{}, err
	}

	// check if user was found from query
	dbUser, err := database.GetUserByEmail(r.DB, signupInput.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &model.AuthResult{}, gqlerror.Errorf("error signing up")
	}
	if dbUser.Email == signupInput.Email {
		return &model.AuthResult{}, gqlerror.Errorf("email already exists")
	}

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupInput.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	verificationCode, err := utils.GenerateVerificationCode(64)
	if err != nil {
		return &model.AuthResult{}, gqlerror.Errorf(err.Error())
	}
	u := database.User{
		Name:               signupInput.Name,
		Email:              signupInput.Email,
		Password:           string(hashedPassword),
		VerificationCode:   verificationCode,
		Verified:           false,
		VerificationSentAt: time.Now(),
	}
	err = r.DB.Create(&u).Error
	if err != nil {
		return &model.AuthResult{}, gqlerror.Errorf(err.Error())
	}

	// should this be moved to inside the user create tx?
	err = mail.SendVerificationCode(verificationCode, u.Email)
	if err != nil {
		return &model.AuthResult{}, gqlerror.Errorf("Issue sending verification email")
	}

	c := &token.Credentials{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}

	refreshToken := token.Sign(c, []byte(os.Getenv(config.REFRESH_SECRET)), config.REFRESH_TTL)
	accessToken := token.Sign(c, []byte(os.Getenv(config.ACCESS_SECRET)), config.ACCESS_TTL)

	return &model.AuthResult{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

// ForgotPassword is the resolver for the forgotPassword field.
func (r *mutationResolver) ForgotPassword(ctx context.Context, email string) (bool, error) {
	panic(fmt.Errorf("not implemented: ForgotPassword - forgotPassword"))
}

// ResendVerificationCode is the resolver for the resendVerificationCode field.
func (r *mutationResolver) ResendVerificationCode(ctx context.Context, email string) (bool, error) {
	verificationCode, err := utils.GenerateVerificationCode(64)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	u := database.User{
		VerificationCode:   verificationCode,
		VerificationSentAt: time.Now(),
	}
	err = database.UpdateUser(r.DB, email, &u)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	// should this be moved to inside the user create tx?
	err = mail.SendVerificationCode(verificationCode, email)
	if err != nil {
		return false, gqlerror.Errorf("Issue sending verification email")
	}

	return true, nil
}

// RefreshAccessToken is the resolver for the refreshAccessToken field.
func (r *mutationResolver) RefreshAccessToken(ctx context.Context, refreshToken string) (*model.RefreshSuccess, error) {
	// read token from context
	claims, err := token.Decode(refreshToken, []byte(os.Getenv(config.REFRESH_SECRET)))
	if err != nil {
		return nil, gqlerror.Errorf("Refresh token invalid")
	}

	accessToken := token.Sign(&token.Credentials{
		ID:    claims.ID,
		Email: claims.Subject,
		Name:  claims.Name,
	},
		[]byte(os.Getenv(config.ACCESS_SECRET)),
		config.ACCESS_TTL,
	)

	return &model.RefreshSuccess{
		AccessToken: accessToken,
	}, nil
}
