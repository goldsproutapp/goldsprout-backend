package util

type HashSet[T comparable] struct {
    hashMap map[T]bool
}

func (s *HashSet[T]) Has(item T) bool {
    _, ok := s.hashMap[item]
    return ok
}

func (s *HashSet[T]) Add(item T) {
    s.hashMap[item] = true
}

func (s *HashSet[T]) Items() []T {
    return MapKeys(s.hashMap)
}

func (s *HashSet[T]) Size() int {
    return len(s.hashMap)
}


func NewHashSet[T comparable]() HashSet[T] {
    return HashSet[T]{hashMap: map[T]bool{}}
}

type OrderedSet[T comparable] struct {
    hashSet HashSet[T]
    order []T
}

func (s *OrderedSet[T]) Has(item T) bool {
    return s.hashSet.Has(item)
}

func (s *OrderedSet[T]) Add(item T) {
    if !s.Has(item) {
        s.order = append(s.order, item)
    }
    s.hashSet.Add(item)
}

func (s *OrderedSet[T]) Items() []T {
    return s.order
}

func NewOrderedSet[T comparable]() OrderedSet[T] {
    return OrderedSet[T]{hashSet: NewHashSet[T](), order: []T{}}
}
