package tristate


type Tristate struct {
    b *bool
}

func BoolPointer(b bool) *bool {
    return &b
}

func True() Tristate {
    return Tristate{b: BoolPointer(true)}
}

func False() Tristate {
    return Tristate{b: BoolPointer(false)}
}

func None() Tristate {
    return Tristate{b: nil}
}

func (t Tristate) IsNone() bool {
    return t.b == nil
}

func (t Tristate) IsTrue() bool {
    return !t.IsNone() && *t.b
}

func (t Tristate) IsFalse() bool {
    return !t.IsNone() && !*t.b
}

func (t Tristate) GetBoolValue(ifNone bool) bool {
    if t.IsNone() {
        return ifNone
    }
    return *t.b
}
