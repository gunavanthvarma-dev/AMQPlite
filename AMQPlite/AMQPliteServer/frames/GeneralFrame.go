package frames

type GeneralFrame interface {
	UnMarshal([]byte) error
	Marshal() ([]byte, error)
}
