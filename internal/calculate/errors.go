package calculate

import "errors"

var (
    ErrInvalidExpression = errors.New("invalid expression")
    ErrDivideByZero      = errors.New("division by zero")
    ErrUnsupportedChar   = errors.New("unsupported character in expression")
)