package errors

import "github.com/pkg/errors"

var (
	ErrCheckingBaseParameters = errors.New("failed to check base parameters: something wrong with ")
	ErrLoadEnvVars            = errors.New("failed to load env vars")
)
