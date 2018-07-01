package transport

const (
	VendorOne  = 0x534c
	ProductOne = 0x0001
	VendorT    = 0x1209
	ProductT   = 0x53C1
)

type Device struct {
	Path      string
	VendorID  int
	ProductID int
}

type Transport interface {
	Write([]byte)
	Read() ([]byte, uint16, int, error)
	Close()
}

type Bus interface {
	Enumerate() ([]Device, error)
	Connect(device Device) (Transport, error)
}
