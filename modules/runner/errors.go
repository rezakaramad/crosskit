package runner

import (
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/crossplane/function-sdk-go/errors"
)

// Fatal marks the function response as fatally failed and returns it.
func Fatal(rsp *fnv1.RunFunctionResponse, msg string, err error) (*fnv1.RunFunctionResponse, error) {
	wrapped := errors.Wrap(err, msg)
	if err == nil {
		wrapped = errors.New(msg)
	}

	response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
		WithMessage(wrapped.Error()).
		TargetCompositeAndClaim()
	response.Fatal(rsp, wrapped)

	return rsp, nil
}
