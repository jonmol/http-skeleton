package dto

import "fmt"

var (
	ErrRude     = fmt.Errorf("no response to rude people")
	ErrVeryRude = fmt.Errorf("outrageous input")
)
