package setters

func NewSkipperSetter() *skipperSetter {
	return &skipperSetter{}
}

// skipperSetter is a setter that skips all operations. Don't rely on this setter
// to get exhausted, since it will never be.
type skipperSetter struct{}

func (s skipperSetter) Init(arg interface{}) error {
	return nil
}

func (s skipperSetter) Execute(op OperationType, value interface{}) error {
	return nil
}

func (s skipperSetter) IsExhausted() bool {
	return false
}

func (s *skipperSetter) GetInner() (Setter, error) {
	return s, nil
}

func (s skipperSetter) setExhaustCallback(eventFunc) {
}

func (s skipperSetter) hasExhaustCallback() bool {
	return true
}

func (s skipperSetter) reset() error {
	return nil
}
