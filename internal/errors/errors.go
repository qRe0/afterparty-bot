package errors

import "github.com/pkg/errors"

var (
	ErrCheckingBaseParameters = errors.New("failed to check base parameters: something wrong with ")
	ErrUpdatingGoogleSheet    = errors.New("failed to update google sheet with transactions")
	ErrGeneratingTicketImg    = errors.New("failed to generate ticket image")
)
