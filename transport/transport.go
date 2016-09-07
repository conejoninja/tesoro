package transport

type Transport interface {
	Write([]byte)
	Read() ([]byte, uint16, int, error)
	Close()
}
