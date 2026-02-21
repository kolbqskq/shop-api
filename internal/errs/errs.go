package errs

import "errors"

var (
	ErrCartNotFound     = errors.New("Корзина не найдена")
	ErrCartAlreadyExist = errors.New("Корзина уже создана")
	ErrInvalidStock     = errors.New("Недопустимое значение Stock")
	ErrInvalidPrice     = errors.New("Недопустимое значение Price")
	ErrNotEnoughStock   = errors.New("Недоступное кол-во")
)
