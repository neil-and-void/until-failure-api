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
	"time"

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
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (model.AuthResult, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, gqlerror.Errorf("Not a valid email")
	}

	dbUser, err := database.GetUserByEmail(r.DB, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gqlerror.Errorf("Email does not exist")
	}
	if err != nil {
		return nil, gqlerror.Errorf("Error Logging In")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
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
func (r *mutationResolver) Signup(ctx context.Context, email string, name string, password string, confirmPassword string) (model.AuthResult, error) {
	if password != confirmPassword {
		return nil, gqlerror.Errorf("Passwords don't match")
	}

	// check strength
	if !utils.IsStrong(password) {
		return nil, gqlerror.Errorf("Password needs at least 1 number and 8 - 16 characters")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, gqlerror.Errorf("Not a valid email")
	}

	// check if user was found from query
	dbUser, err := database.GetUserByEmail(r.DB, email)
	if dbUser.ID != 0 {
		return nil, gqlerror.Errorf("Email already exists")
	}

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	u := database.User{Name: name, Email: email, Password: string(hashedPassword)}
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

// CreateWorkoutRoutine is the resolver for the createWorkoutRoutine field.
func (r *mutationResolver) CreateWorkoutRoutine(ctx context.Context, routine model.WorkoutRoutineInput) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, err
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
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Creating Workout Routine")
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
		Active:           wr.Active,
	}, nil
}

// UpdateWorkoutRoutine is the resolver for the updateWorkoutRoutine field.
func (r *mutationResolver) UpdateWorkoutRoutine(ctx context.Context, workoutRoutine model.UpdateWorkoutRoutineInput) (*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutine.ID)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Updating Workout Routine: Access Denied")
	}

	var exerciseRoutines []*database.ExerciseRoutine
	for _, er := range workoutRoutine.ExerciseRoutines {
		// newly added exercises won't have an ID
		// nil ID indicates that this exercise should be created, otherwise update
		// the exercise that has that ID
		var model gorm.Model
		if er.ID != nil {
			num, err := strconv.ParseUint(*er.ID, 10, strconv.IntSize)
			if err != nil {
				panic(err)
			}
			model.ID = uint(num)
		}

		workoutRoutineIDUint, err := strconv.ParseUint(workoutRoutine.ID, 10, strconv.IntSize)
		if err != nil {
			panic(err)
		}

		exerciseRoutines = append(exerciseRoutines, &database.ExerciseRoutine{
			Model:            model,
			Name:             er.Name,
			Sets:             uint(er.Sets),
			Reps:             uint(er.Reps),
			WorkoutRoutineID: uint(workoutRoutineIDUint),
		})
	}

	err = database.UpdateWorkoutRoutine(r.DB, workoutRoutine.ID, workoutRoutine.Name, exerciseRoutines)
	if err != nil {
		return &model.WorkoutRoutine{}, gqlerror.Errorf("Error Updating Workout Routine")
	}

	var updatedExerciseRoutines []*model.ExerciseRoutine
	for _, er := range exerciseRoutines {
		updatedExerciseRoutines = append(updatedExerciseRoutines, &model.ExerciseRoutine{
			ID:   fmt.Sprintf("%d", er.ID),
			Name: er.Name,
			Reps: int(er.Reps),
			Sets: int(er.Sets),
		})
	}

	return &model.WorkoutRoutine{
		ID:               workoutRoutine.ID,
		Name:             workoutRoutine.Name,
		ExerciseRoutines: updatedExerciseRoutines,
	}, nil
}

// DeleteWorkoutRoutine is the resolver for the deleteWorkoutRoutine field.
func (r *mutationResolver) DeleteWorkoutRoutine(ctx context.Context, workoutRoutineID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Routine: Access Denied")
	}

	err = database.DeleteWorkoutRoutine(r.DB, workoutRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Routine")
	}

	return 1, nil
}

// AddExerciseRoutine is the resolver for the addExerciseRoutine field.
func (r *mutationResolver) AddExerciseRoutine(ctx context.Context, workoutRoutineID string, exerciseRoutine model.ExerciseRoutineInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine: Access Denied")
	}

	workoutRoutineIDUint, err := strconv.ParseUint(workoutRoutineID, 10, strconv.IntSize)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine")
	}
	dbExerciseRoutine := &database.ExerciseRoutine{
		Name:             exerciseRoutine.Name,
		Sets:             uint(exerciseRoutine.Sets),
		Reps:             uint(exerciseRoutine.Reps),
		WorkoutRoutineID: uint(workoutRoutineIDUint),
	}
	err = database.AddExerciseRoutine(r.DB, dbExerciseRoutine)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise Routine")
	}

	return fmt.Sprintf("%d", dbExerciseRoutine.ID), nil
}

// DeleteExerciseRoutine is the resolver for the deleteExerciseRoutine field.
func (r *mutationResolver) DeleteExerciseRoutine(ctx context.Context, exerciseRoutineID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	exerciseRoutine := database.ExerciseRoutine{}
	err = database.GetExerciseRoutine(r.DB, exerciseRoutineID, &exerciseRoutine)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine")
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, fmt.Sprintf("%d", exerciseRoutine.WorkoutRoutineID))
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine: Access Denied")
	}

	err = database.DeleteExerciseRoutine(r.DB, exerciseRoutineID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise Routine")
	}

	return 1, nil
}

// AddWorkoutSession is the resolver for the addWorkoutSession field.
func (r *mutationResolver) AddWorkoutSession(ctx context.Context, workout model.WorkoutSessionInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	var dbExercises []database.Exercise
	for _, e := range workout.Exercises {
		var set []database.SetEntry

		for _, s := range e.SetEntries {
			set = append(set, database.SetEntry{
				Weight: float32(s.Weight),
				Reps:   uint(s.Reps),
			})
		}

		exerciseRoutineId, err := strconv.ParseUint(e.ExerciseRoutineID, 10, 32)
		if err != nil {
			return "", gqlerror.Errorf("Error Adding Workout Session")
		}

		dbExercises = append(dbExercises, database.Exercise{
			Sets:              set,
			ExerciseRoutineID: uint(exerciseRoutineId),
			Notes:             e.Notes,
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

// UpdateWorkoutSession is the resolver for the updateWorkoutSession field.
func (r *mutationResolver) UpdateWorkoutSession(ctx context.Context, workoutSessionID string, updateWorkoutSessionInput model.UpdateWorkoutSessionInput) (*model.UpdatedWorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.UpdatedWorkoutSession{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return &model.UpdatedWorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session: Access Denied")
	}

	var start time.Time
	if updateWorkoutSessionInput.Start != nil {
		start = *updateWorkoutSessionInput.Start
	}
	updatedWorkoutSession := database.WorkoutSession{
		Start: start,
		End:   updateWorkoutSessionInput.End,
	}
	err = database.UpdateWorkoutSession(r.DB, workoutSessionID, &updatedWorkoutSession)
	if err != nil {
		return &model.UpdatedWorkoutSession{}, gqlerror.Errorf("Error Updating Workout Session")
	}

	return &model.UpdatedWorkoutSession{
		ID:    fmt.Sprintf("%d", updatedWorkoutSession.ID),
		Start: updatedWorkoutSession.Start,
		End:   updatedWorkoutSession.End,
	}, nil
}

// DeleteWorkoutSession is the resolver for the deleteWorkoutSession field.
func (r *mutationResolver) DeleteWorkoutSession(ctx context.Context, workoutSessionID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Session: Access Denied")
	}

	err = database.DeleteWorkoutSession(r.DB, workoutSessionID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Workout Session")
	}

	return 1, nil
}

// AddExercise is the resolver for the addExercise field.
func (r *mutationResolver) AddExercise(ctx context.Context, workoutSessionID string, exercise model.ExerciseInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutSession(userId, workoutSessionID)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	// todo: check can access exercise routines that are being added

	var setEntries []database.SetEntry
	for _, s := range exercise.SetEntries {
		setEntries = append(setEntries, database.SetEntry{
			Reps:   uint(s.Reps),
			Weight: float32(s.Weight),
		})
	}

	workoutSessionIDUint, err := strconv.ParseUint(workoutSessionID, 10, 32)
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
		Notes:             exercise.Notes,
	}

	err = database.AddExercise(r.DB, dbExercise, workoutSessionID)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Exercise: %s", err.Error())
	}

	return fmt.Sprintf("%d", dbExercise.ID), nil
}

// UpdateExercise is the resolver for the updateExercise field.
func (r *mutationResolver) UpdateExercise(ctx context.Context, exerciseID string, exercise model.UpdateExerciseInput) (*model.UpdatedExercise, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.UpdatedExercise{}, err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, strconv.IntSize)
	dbExercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &dbExercise)
	if err != nil {
		return &model.UpdatedExercise{}, gqlerror.Errorf("Error Updating Exercise")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", dbExercise.WorkoutSessionID))
	if err != nil {
		return &model.UpdatedExercise{}, gqlerror.Errorf("Error Updating Exercise: Access Denied")
	}

	updatedExercise := database.Exercise{
		Notes: exercise.Notes,
	}
	err = database.UpdateExercise(r.DB, exerciseID, &updatedExercise)
	if err != nil {
		return &model.UpdatedExercise{}, gqlerror.Errorf("Error Updating Exercise")
	}

	return &model.UpdatedExercise{
		ID:    exerciseID,
		Notes: updatedExercise.Notes,
	}, nil
}

// DeleteExercise is the resolver for the deleteExercise field.
func (r *mutationResolver) DeleteExercise(ctx context.Context, exerciseID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, strconv.IntSize)
	dbExercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &dbExercise)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", dbExercise.WorkoutSessionID))
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise: Access Denied")
	}

	err = database.DeleteExercise(r.DB, exerciseID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Exercise")
	}

	return 1, nil
}

// AddSet is the resolver for the addSet field.
func (r *mutationResolver) AddSet(ctx context.Context, exerciseID string, set *model.SetEntryInput) (string, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return "", err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set: Invalid Exercise ID")
	}
	exercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set %s", err)
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set: Access Denied")
	}

	dbSet := database.SetEntry{
		ExerciseID: uint(exerciseIDUint),
		Weight:     float32(set.Weight),
		Reps:       uint(set.Reps),
	}
	err = database.AddSet(r.DB, &dbSet)
	if err != nil {
		return "", gqlerror.Errorf("Error Adding Set")
	}

	return fmt.Sprintf("%d", dbSet.ID), nil
}

// UpdateSet is the resolver for the updateSet field.
func (r *mutationResolver) UpdateSet(ctx context.Context, setID string, set model.UpdateSetEntryInput) (*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.SetEntry{}, err
	}

	var setEntry database.SetEntry
	err = database.GetSet(r.DB, &setEntry, setID)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	exercise := database.Exercise{
		Model: gorm.Model{
			ID: setEntry.ExerciseID,
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set: Access Denied")
	}

	// check optional inputs
	var reps uint
	if set.Reps != nil {
		reps = uint(*set.Reps)
	}
	var weight float32
	if set.Weight != nil {
		weight = float32(*set.Weight)
	}

	updatedSet := database.SetEntry{
		Reps:   reps,
		Weight: weight,
	}
	err = database.UpdateSet(r.DB, setID, &updatedSet)
	if err != nil {
		return &model.SetEntry{}, gqlerror.Errorf("Error Updating Set")
	}

	return &model.SetEntry{
		ID:     fmt.Sprintf("%d", updatedSet.ID),
		Weight: float64(updatedSet.Weight),
		Reps:   int(updatedSet.Reps),
	}, nil
}

// DeleteSet is the resolver for the deleteSet field.
func (r *mutationResolver) DeleteSet(ctx context.Context, setID string) (int, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return 0, err
	}

	var setEntry database.SetEntry
	err = database.GetSet(r.DB, &setEntry, setID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	exercise := database.Exercise{
		Model: gorm.Model{
			ID: setEntry.ExerciseID,
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set: Access Denied")
	}

	err = database.DeleteSet(r.DB, setID)
	if err != nil {
		return 0, gqlerror.Errorf("Error Deleting Set")
	}

	return 1, nil
}

// WorkoutRoutines is the resolver for the workoutRoutines field.
func (r *queryResolver) WorkoutRoutines(ctx context.Context) ([]*model.WorkoutRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.WorkoutRoutine{}, err
	}

	dbwr, err := database.GetWorkoutRoutines(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return []*model.WorkoutRoutine{}, gqlerror.Errorf("Error Getting Workout Routine")
	}

	// map database workout routine to graphql workout routine
	var workoutRoutines []*model.WorkoutRoutine
	for _, wr := range dbwr {
		// map database exercise routine to graphql exercise routine
		var exerciseRoutines []*model.ExerciseRoutine
		for _, er := range wr.ExerciseRoutines {
			exerciseRoutines = append(exerciseRoutines, &model.ExerciseRoutine{
				ID:     fmt.Sprintf("%d", er.ID),
				Name:   er.Name,
				Active: er.Active,
				Sets:   int(er.Sets),
				Reps:   int(er.Reps),
			})
		}
		workoutRoutines = append(workoutRoutines, &model.WorkoutRoutine{
			ID:               fmt.Sprintf("%d", wr.ID),
			Name:             wr.Name,
			Active:           wr.Active,
			ExerciseRoutines: exerciseRoutines,
		})
	}
	return workoutRoutines, nil
}

// ExerciseRoutines is the resolver for the exerciseRoutines field.
func (r *queryResolver) ExerciseRoutines(ctx context.Context, workoutRoutineID string) ([]*model.ExerciseRoutine, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.ExerciseRoutine{}, err
	}

	userId := fmt.Sprintf("%d", u.ID)
	err = r.ACS.CanAccessWorkoutRoutine(userId, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine: Access Denied")
	}

	erdb, err := database.GetExerciseRoutines(r.DB, workoutRoutineID)
	if err != nil {
		return []*model.ExerciseRoutine{}, gqlerror.Errorf("Error Getting Exercise Routine")
	}

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
		return []*model.WorkoutSession{}, err
	}

	dbWorkoutSessions, err := database.GetWorkoutSessions(r.DB, fmt.Sprintf("%d", u.ID))
	if err != nil {
		return []*model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Sessions")
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
				})

			}

			exercise = append(exercise, &model.Exercise{
				ID:                fmt.Sprintf("%d", e.ID),
				Sets:              setEntries,
				Notes:             e.Notes,
				ExerciseRoutineID: fmt.Sprintf("%d", e.ExerciseRoutineID),
			})
		}

		workoutSessions = append(workoutSessions, &model.WorkoutSession{
			ID:               fmt.Sprintf("%d", ws.ID),
			Start:            ws.Start,
			End:              ws.End,
			WorkoutRoutineID: fmt.Sprintf("%d", ws.WorkoutRoutineID),
			Exercises:        exercise,
		})
	}

	return workoutSessions, nil
}

// WorkoutSession is the resolver for the workoutSession field.
func (r *queryResolver) WorkoutSession(ctx context.Context, workoutSessionID string) (*model.WorkoutSession, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.WorkoutSession{}, err
	}

	var dbWorkoutSession database.WorkoutSession
	err = database.GetWorkoutSession(r.DB, fmt.Sprintf("%d", u.ID), workoutSessionID, &dbWorkoutSession)
	if err != nil {
		return &model.WorkoutSession{}, gqlerror.Errorf("Error Getting Workout Session: Access Denied")
	}

	var exercises []*model.Exercise
	for _, e := range dbWorkoutSession.Exercises {

		var setEntries []*model.SetEntry
		for _, s := range e.Sets {
			setEntries = append(setEntries, &model.SetEntry{
				ID:     fmt.Sprintf("%d", s.ID),
				Weight: float64(s.Weight),
				Reps:   int(s.Reps),
			})

		}

		exercises = append(exercises, &model.Exercise{
			ID:                fmt.Sprintf("%d", e.ID),
			ExerciseRoutineID: fmt.Sprintf("%d", e.ExerciseRoutineID),
			Sets:              setEntries,
			Notes:             e.Notes,
		})
	}

	return &model.WorkoutSession{
		ID:               fmt.Sprintf("%d", dbWorkoutSession.ID),
		Start:            dbWorkoutSession.Start,
		End:              dbWorkoutSession.End,
		WorkoutRoutineID: fmt.Sprintf("%d", dbWorkoutSession.WorkoutRoutineID),
		Exercises:        exercises,
	}, nil
}

// Exercise is the resolver for the exercise field.
func (r *queryResolver) Exercise(ctx context.Context, exerciseID string) (*model.Exercise, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return &model.Exercise{}, err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return &model.Exercise{}, gqlerror.Errorf("Error Getting Exercise: Invalid Exercise ID")
	}

	exercise := &database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, exercise)
	if err != nil {
		return &model.Exercise{}, gqlerror.Errorf("Error Getting Exercise: %s", err.Error())
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return &model.Exercise{}, gqlerror.Errorf("Error Getting Exercise: %s", err.Error())
	}

	var setEntries []*model.SetEntry
	for _, s := range exercise.Sets {
		setEntries = append(setEntries, &model.SetEntry{
			ID:     fmt.Sprintf("%d", s.ID),
			Weight: float64(s.Weight),
			Reps:   int(s.Reps),
		})
	}

	return &model.Exercise{
		ID:                exerciseID,
		Sets:              setEntries,
		Notes:             exercise.Notes,
		ExerciseRoutineID: fmt.Sprintf("%d", exercise.ExerciseRoutineID),
	}, nil
}

// Exercises is the resolver for the exercises field.
func (r *queryResolver) Exercises(ctx context.Context, workoutSessionID string) ([]*model.Exercise, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.Exercise{}, err
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), workoutSessionID)
	if err != nil {
		return []*model.Exercise{}, gqlerror.Errorf("Error Getting Exercises: %s", err.Error())
	}

	var dbExercises []database.Exercise
	err = database.GetExercises(r.DB, &dbExercises, workoutSessionID)
	if err != nil {
		return []*model.Exercise{}, gqlerror.Errorf("Error Getting Exercises")
	}

	var exercises []*model.Exercise
	for _, e := range dbExercises {

		var setEntries []*model.SetEntry
		for _, s := range e.Sets {
			setEntries = append(setEntries, &model.SetEntry{
				ID:     fmt.Sprintf("%d", s.ID),
				Weight: float64(s.Weight),
				Reps:   int(s.Reps),
			})
		}

		exercises = append(exercises, &model.Exercise{
			ID:                fmt.Sprintf("%d", e.ID),
			Sets:              setEntries,
			Notes:             e.Notes,
			ExerciseRoutineID: fmt.Sprintf("%d", e.ExerciseRoutineID),
		})
	}
	return exercises, nil
}

// Sets is the resolver for the sets field.
func (r *queryResolver) Sets(ctx context.Context, exerciseID string) ([]*model.SetEntry, error) {
	u, err := middleware.GetUser(ctx)
	if err != nil {
		return []*model.SetEntry{}, err
	}

	exerciseIDUint, err := strconv.ParseUint(exerciseID, 10, 64)
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets: Invalid Exercise ID")
	}
	exercise := database.Exercise{
		Model: gorm.Model{
			ID: uint(exerciseIDUint),
		},
	}
	err = database.GetExercise(r.DB, &exercise)
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets")
	}

	err = r.ACS.CanAccessWorkoutSession(fmt.Sprintf("%d", u.ID), fmt.Sprintf("%d", exercise.WorkoutSessionID))
	if err != nil {
		return []*model.SetEntry{}, gqlerror.Errorf("Error Getting Sets: Access Denied")
	}

	var sets []*model.SetEntry
	for _, s := range exercise.Sets {
		sets = append(sets, &model.SetEntry{
			ID:     fmt.Sprintf("%d", s.ID),
			Reps:   int(s.Reps),
			Weight: float64(s.Weight),
		})
	}

	return sets, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
