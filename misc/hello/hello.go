//gp:build protogo
package hello

type ProtoGoHello struct {
	// Exported types must have a protogo field tag
	Message string `protogo:"1"`
	Other   string `protogo:"2"`

	// Non-exported types don't need a protogo field tag
	id int64
}

type GlobalType struct {
	Data string `protogo:"1"`
}

var Global GlobalType

func Hello(_ GlobalType) ProtoGoHello {
	return ProtoGoHello{
		Message: "Hello",
		id:      42,
	}
}
