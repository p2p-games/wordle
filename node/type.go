package node

// Type defines the Node type for identity purposes.
// The zero value for Type is invalid.
type Type uint8

const (
	// Light is a stripped-down Wordle Node which aims to be lightweight while preserving the highest possible
	// security guarantees.
	Light Type = iota + 1
	// Full is a Wordle Node that stores blocks in their entirety.
	Full
)

// String converts Type to its string representation.
func (t Type) String() string {
	if !t.IsValid() {
		return "unknown"
	}
	return typeToString[t]
}

// IsValid reports whether the Type is valid.
func (t Type) IsValid() bool {
	_, ok := typeToString[t]
	return ok
}

// ParseType converts string in a type if possible.
func ParseType(str string) Type {
	tp, ok := stringToType[str]
	if !ok {
		return 0
	}

	return tp
}

// typeToString keeps string representations of all valid Types.
var typeToString = map[Type]string{
	Light: "Light",
	Full:  "Full",
}

// typeToString maps strings representations of all valid Types.
var stringToType = map[string]Type{
	"Light": Light,
	"Full":  Full,
}
