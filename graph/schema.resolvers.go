package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strconv"

	"github.com/neilZon/workout-logger-api/config"
	"github.com/neilZon/workout-logger-api/database"
	"github.com/neilZon/workout-logger-api/graph/generated"
	"github.com/neilZon/workout-logger-api/graph/model"
	"github.com/neilZon/workout-logger-api/middleware"
	"github.com/neilZon/workout-logger-api/token"
	"github.com/neilZon/workout-logger-api/utils"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, email *string, password *string) (model.AuthResult, error) {
	if _, err := mail.ParseAddress(*email); err != nil {
		return nil, gqlerror.Errorf("Not a valid email")
	}

	dbUser, err := database.GetUserByEmail(r.DB, *email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gqlerror.Errorf("Email does not exist")
	}
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(*password)); err != nil {
		return nil, gqlerror.Errorf("Incorrect Password")
	}
	c := &token.Credentials{
		ID:    dbUser.ID,
		Email: dbUser.Email,
		Name:  dbUser.Name,
	}

	refreshToken := token.Sign(c, []byte(os.Getenv(config.REFRESH_SECRET)), config.REFRESH_TTL)
	accessToken := token.Sign(c, []byte(os.Getenv(config.ACCESS_SECRET)), config.ACCESS_TTL)

	return model.AuthSuccess{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

// Signup is the resolver for the signup field.
func (r *mutationResolver) Signup(ctx context.Context, email *string, name *string, password *string, confirmPassword *string) (model.AuthResult, error) {
	if *password != *confirmPassword {
		return nil, gqlerror.Errorf("Passwords don't match")
	}

	// check strength
	if !utils.IsStrong(*password) {
		return nil, gqlerror.Errorf("Password needs at least 1 number and 8 - 16 characters")
	}

	if _, err := mail.ParseAddress(*email); err != nil {
		return nil, gqlerror.Errorf("Not a valid email")
	}

	// check if user was found from query
	dbUser, err := database.GetUserByEmail(r.DB, *email)
	if dbUser.ID != 0 {
		return nil, gqlerror.Errorf("Email already exists")
	}

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	u := database.User{Name: *name, Email: *email, Password: string(hashedPassword)}
	err = r.DB.Create(&u).Error
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	c := &token.Credentials{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}

	refreshToken := token.Sign(c, []byte(os.Getenv(config.REFRESH_SECRET)), config.REFRESH_TTL)
	accessToken := token.Sign(c, []byte(os.Getenv(config.ACCESS_SECRET)), config.ACCESS_TTL)

	return model.AuthSuccess{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

// RefreshAccessToken is the resolver for the refreshAccessToken field.
func (r *mutationResolver) RefreshAccessToken(ctx context.Context, refreshToken *string) (*model.RefreshSuccess, error) {
	// read token from context
	claims, err := token.Decode(*refreshToken, []byte(os.Getenv(config.REFRESH_SECRET)))
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

// CreateWorkoutRoutine is the resolver for the createWorkoutRoutine field.
func (r *mutationResolver) CreateWorkoutRoutine(ctx context.Context, routine *model.WorkoutRoutineInput) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Creating Workout: %s", err.Error())
	}

	// validate input
	if len([]rune(routine.Name)) <= 2 {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Invalid Routine Name Length")
	}

	exerciseRoutines := make([]database.ExerciseRoutine, 0)
	for _, er := range routine.ExerciseRoutines {
		exerciseRoutines = append(exerciseRoutines, database.ExerciseRoutine{Name: er.Name, Reps: uint(er.Reps), Sets: uint(er.Sets)})
	}

	wr := &database.WorkoutRoutine{
		Name:             routine.Name,
		ExerciseRoutines: exerciseRoutines,
		UserID:           u.ID,
	}

	res := database.CreateWorkoutRoutine(r.DB, wr)
	if res.Error != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf(err.Error())
	}

	dbExerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range wr.ExerciseRoutines {
		dbExerciseRoutines = append(dbExerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Sets: int(er.Sets),
			Reps: int(er.Reps),
		})
	}

	return &model.WorkoutRoutine{
		ID:               fmt.Sprintf("%d", wr.ID),
		Name:             wr.Name,
		ExerciseRoutines: []*model.ExerciseRoutine{},
	}, nil
}

// AddWorkoutSession is the resolver for the addWorkoutSession field.
func (r *mutationResolver) AddWorkoutSession(ctx context.Context, workout *model.WorkoutSessionInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Workout Session: Invalid Token")
	}

	var dbExercises []database.Exercise
	for _, e := range workout.Exercises {
		var set []database.SetEntry

		for _, s := range e.SetEntries {
			set = append(set, database.SetEntry{
				Weight: float32(s.Weight),
				Reps:   uint(s.Reps),
				Notes:  s.Notes,
			})
		}

		exerciseRoutineId, err := strconv.ParseUint(e.ExerciseRoutineID, 10, 32)
		if err != nil {
			return "", gqlerror.Errorf("Error Adding Workout Session")
		}

		dbExercises = append(dbExercises, database.Exercise{
			Sets:              set,
			ExerciseRoutineID: uint(exerciseRoutineId),
		})
	}

	workotuRoutineID, err := strconv.ParseUint(workout.WorkoutRoutineID, 10, 64)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Workout Session: Invalid Workout Routine ID")
	}

	ws := &database.WorkoutSession{
		Start:            workout.Start,
		End:              workout.End,
		WorkoutRoutineID: uint(workotuRoutineID),
		UserID:           u.ID,
		Exercises:        dbExercises,
	}
	err = database.AddWorkoutSession(r.DB, ws)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Workout Session")
	}

	return fmt.Sprintf("%d", ws.ID), nil
}

// AddExercise is the resolver for the addExercise field.
func (r *mutationResolver) AddExercise(ctx context.Context, exercise *model.ExerciseInput, workoutSessionID *string) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.AC.CanAccessWorkoutSession(userId, *workoutSessionID)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	var setEntries []database.SetEntry
	for _, s := range exercise.SetEntries {
		setEntries = append(setEntries, database.SetEntry{
			Reps: uint(s.Reps),
			Weight: float32(s.Weight),
			Notes: s.Notes,
		})
	}

	workoutSessionIDUint, err := strconv.ParseUint(*workoutSessionID, 10, 32)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	exerciseRoutineID, err := strconv.ParseUint(exercise.ExerciseRoutineID, 10, 32)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	dbExercise := &database.Exercise{
		WorkoutSessionID:  uint(workoutSessionIDUint),
		ExerciseRoutineID: uint(exerciseRoutineID),
		Sets:              setEntries,
	}
	database.AddExercise(r.DB, dbExercise, *workoutSessionID)

	return fmt.Sprintf("%d", dbExercise.ID), nil
}

// WorkoutRoutines is the resolver for the workoutRoutines field.
func (r *queryResolver) WorkoutRoutines(ctx context.Context) ([]*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.WorkoutRoutine{}, gqlerror.Errorf("Error Getting Workout Routine: %s", err.Error())
	}

	dbwr, err := database.GetWorkoutRoutines(r.DB, u.Subject)
	if err != nil {
		return []*model.WorkoutRoutine{}, gqlerror.Errorf("Error Getting Workout Routine")
	}

	// map database workout routine to graphql workout routine
	workoutRoutines := make([]*model.WorkoutRoutine, 0)
	for _, wr := range dbwr {

		// map database exercise routine to graphql exercise routine
		exerciseRoutines := make([]*model.ExerciseRoutine, 0)
		for _, er := range wr.ExerciseRoutines {
			exerciseRoutines = append(exerciseRoutines, &model.ExerciseRoutine{
				ID:   fmt.Sprintf("%d", er.ID),
				Name: er.Name,
				Sets: int(er.Sets),
				Reps: int(er.Reps),
			})
		}
		workoutRoutines = append(workoutRoutines, &model.WorkoutRoutine{
			ID:               fmt.Sprintf("%d", wr.ID),
			Name:             wr.Name,
			ExerciseRoutines: exerciseRoutines,
		})
	}
	return workoutRoutines, nil
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *queryResolver) ExerciseRoutines(ctx context.Context, workoutRoutineID *string) ([]*model.ExerciseRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Workout Routine: %s", err.Error())
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.AC.CanAccessWorkoutRoutine(userId, *workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Workout Routine: %s", err.Error())
	}

	erdb, err := database.GetExerciseRoutines(r.DB, *workoutRoutineID)

	exerciseRoutines := make([]*model.ExerciseRoutine, 0)
	for _, er := range erdb {
		exerciseRoutines = append(exerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Sets: int(er.Sets),
			Reps: int(er.Reps),
		})
	}

	return exerciseRoutines, nil
}

// WorkoutSessions is the resolver for the workoutSessions field.
func (r *queryResolver) WorkoutSessions(ctx context.Context) ([]*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Routine: %s", err.Error())
	}

	dbWorkoutSessions, err := database.GetWorkoutSessions(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return []*model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Sessions: %s", err.Error())
	}

	var workoutSessions []*model.WorkoutSession
	for _, ws := range dbWorkoutSessions {

		var exercise []*model.Exercise
		for _, e := range ws.Exercises {

			var setEntries []*model.SetEntry
			for _, s := range e.Sets {
				setEntries = append(setEntries, &model.SetEntry{
					ID:     fmt.Sprintf("%d", s.ID),
					Weight: float64(s.Weight),
					Reps:   int(s.Reps),
					Notes:  s.Notes,
				})

			}

			exercise = append(exercise, &model.Exercise{
				ID:   fmt.Sprintf("%d", e.ID),
				Sets: setEntries,
			})
		}

		workoutSessions = append(workoutSessions, &model.WorkoutSession{
			ID:       fmt.Sprintf("%d", ws.ID),
			Start:    ws.Start,
			End:      ws.End,
			Exercise: exercise,
		})
	}

	return workoutSessions, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
