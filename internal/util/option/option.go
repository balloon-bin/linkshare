package option

type Option[T any] struct {
	hasValue bool
	value    T
}

func Some[T any](value T) Option[T] {
	return Option[T]{
		hasValue: true,
		value:    value,
	}
}

func None[T any]() Option[T] {
	return Option[T]{
		hasValue: false,
	}
}

func (o Option[T]) IsSome() bool {
	return o.hasValue
}

func (o Option[T]) IsNone() bool {
	return !o.hasValue
}

func (o Option[T]) Value() T {
	if !o.hasValue {
		panic("Option has no value")
	}
	return o.value
}

func (o Option[T]) ValueOr(defaultValue T) T {
	if !o.hasValue {
		return defaultValue
	}
	return o.value
}
