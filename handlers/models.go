package handlers

type (
	User struct {
		Name string `validate:"required,max=50"`
	}
)
