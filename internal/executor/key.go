package executor

type Key[T any] struct {
	name string
}

func NewKey[T any](name string) Key[T] {
	return Key[T]{name: name}
}

func (k Key[T]) Name() string {
	return k.name
}

var (
	ZoneIDKey     = NewKey[string]("zoneID")
	ZoneNameKey   = NewKey[string]("zoneName")
	RecordIDKey   = NewKey[string]("recordID")
	RecordNameKey = NewKey[string]("recordName")
)
