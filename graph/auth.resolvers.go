package graph

import (
	"context"
	"errors"
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
	err := validator.ValidateEmail(loginInput.Email)
	if err != nil {
		return &model.AuthResult{}, gqlerror.Errorf("invalid email")
	}

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
	now := time.Now()
	u := database.User{
		Name:               signupInput.Name,
		Email:              signupInput.Email,
		Password:           string(hashedPassword),
		VerificationCode:   &verificationCode,
		Verified:           false,
		VerificationSentAt: &now,
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

// ResendVerificationCode is the resolver for the resendVerificationCode field.
func (r *mutationResolver) ResendVerificationCode(ctx context.Context, email string) (bool, error) {
	err := validator.ValidateEmail(email)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	// check if user exists to send email to
	_, err = database.GetUserByEmail(r.DB, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, gqlerror.Errorf("user does not exist")
	}
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	verificationCode, err := utils.GenerateVerificationCode(64)
	if err != nil {
		return false, gqlerror.Errorf("could not send verification email")
	}

	now := time.Now()
	u := database.User{
		VerificationCode:   &verificationCode,
		VerificationSentAt: &now,
	}
	err = database.UpdateUser(r.DB, email, &u)
	if err != nil {
		return false, gqlerror.Errorf("could not send verification email")
	}

	// should this be moved to inside the user create tx?
	err = mail.SendVerificationCode(verificationCode, email)
	if err != nil {
		return false, gqlerror.Errorf("could not send verification email")
	}

	return true, nil
}

// SendForgotPasswordLink is the resolver for the sendForgotPasswordLink field.
func (r *mutationResolver) SendForgotPasswordLink(ctx context.Context, email string) (bool, error) {
	err := validator.ValidateEmail(email)
	if err != nil {
		return false, gqlerror.Errorf("not a valid email")
	}

	// check if user exists to send email to
	_, err = database.GetUserByEmail(r.DB, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, gqlerror.Errorf("user does not exist")
	}
	if err != nil {
		return false, gqlerror.Errorf("error sending password reset code")
	}

	passwordResetCode, err := utils.GenerateVerificationCode(64)
	if err != nil {
		return false, gqlerror.Errorf("error sending password reset code")
	}

	now := time.Now()
	u := database.User{
		PasswordResetCode:   &passwordResetCode,
		PasswordResetSentAt: &now,
	}
	err = database.UpdateUser(r.DB, email, &u)
	if err != nil {
		return false, gqlerror.Errorf("error sending password reset code")
	}

	err = mail.SendResetLink(passwordResetCode, email)
	if err != nil {
		return false, gqlerror.Errorf("error sending password reset code")
	}

	return true, nil
}

// ResetPassword is the resolver for the resetPassword field.
func (r *mutationResolver) ResetPassword(ctx context.Context, passwordResetCredentials model.PasswordResetCredentials) (bool, error) {
	if passwordResetCredentials.Password != passwordResetCredentials.ConfirmPassword {
		return false, gqlerror.Errorf("passwords don't match")
	}

	user, err := database.GetUserByPasswordCode(r.DB, passwordResetCredentials.Code)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}
	expiryTime := time.Now().Add(24 * time.Hour)
	if user.PasswordResetCode == nil || *user.PasswordResetCode != passwordResetCredentials.Code || user.PasswordResetSentAt == nil || user.PasswordResetSentAt.After(expiryTime) {
		return false, gqlerror.Errorf("could not reset password")
	}

	// Hashing the password with the default cost of 10
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordResetCredentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return false, gqlerror.Errorf("could not reset password")
	}

	err = database.ChangePassword(r.DB, passwordResetCredentials.Code, string(newHashedPassword))
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}
