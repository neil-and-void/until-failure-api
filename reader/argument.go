package reader

import (
	"fmt"
	"strings"
	"time"
)

const (
	workoutSessionId = 0
	workoutRoutineId = 1
	date = 2
)

// serializable struct that can be passed to the Prev Exercise loader function
type PrevExerciseArgs struct {
	WorkoutSessionId string
	WorkoutRoutineID string
	Date time.Time
}

func (p *PrevExerciseArgs) String() string {
	dateString := p.Date.Format(time.RFC3339)	
	// serializes into comma separated string
	return fmt.Sprintf("%s,%s,%s", p.WorkoutSessionId, p.WorkoutRoutineID, dateString)
}

func BuildPrevExerciseArgs(s string) (*PrevExerciseArgs, error) {
	args := strings.Split(s, ",")
	date, err := time.Parse(time.RFC3339, args[date]) 
	if err != nil {
		panic(err)
	}

	return &PrevExerciseArgs{
		WorkoutSessionId: args[workoutSessionId],
		WorkoutRoutineID: args[workoutRoutineId],
		Date: date,
	}, nil
}
