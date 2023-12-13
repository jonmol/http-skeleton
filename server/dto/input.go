package dto

// parameters for the Hello endpoint
type InputHello struct {
	Input string `json:"input" validate:"required,max=200"`
}
