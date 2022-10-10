package graph

import (
	"github.com/neilZon/workout-logger-api/accesscontrol"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *gorm.DB
	AC *accesscontrol.AccessController
}
