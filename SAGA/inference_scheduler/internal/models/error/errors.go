package error

import "errors"

var (
	// ErrDuplicateKey возвращается когда происходит нарушение уникального ключа (duplicate key)
	// Это нормальная ситуация при повторной обработке сообщения (идемпотентность)
	ErrDuplicateKey = errors.New("duplicate key: record already exists")
)
