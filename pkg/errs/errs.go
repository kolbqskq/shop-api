package errs

import "errors"

var (
	ErrCartNotFound     = errors.New("Корзина не найдена")
	ErrCartAlreadyExist = errors.New("Корзина уже создана")
)
