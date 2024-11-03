package util

type TernaryAssignment[T any] struct {
	condition bool
	ifTrue    T
	ifFalse   T
}

func Assign[T any](ifTrue T) TernaryAssignment[T] {
    return TernaryAssignment[T]{
        ifTrue: ifTrue,
    }
}

func (t TernaryAssignment[T]) If(cond bool) TernaryAssignment[T] {
    t.condition = cond
    return t
}

func (t TernaryAssignment[T]) Else(ifFalse T) T {
    t.ifFalse = ifFalse
    return t.Evaluate()
}

func (t TernaryAssignment[T]) Evaluate() T {
    if t.condition {
        return t.ifTrue
    }
    return t.ifFalse
}
