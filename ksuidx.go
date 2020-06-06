package ksuidx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
)

var (
	// Unknown namespace will be applied when converting from raw ksuid.KSUID values
	Unknown = MustNamespace("unk")
	Nil     = ID{
		ns:    Unknown,
		ksuid: ksuid.Nil,
	}
)

var (
	errNamespaceSize = fmt.Errorf("valid namespaces are %v bytes", nsLength)
	errStringSize    = fmt.Errorf("valid encoded strings are %v bytes", stringEncodedLength)
)

const (
	byteLength          = 23 // length of ksuidx id
	baseLength          = 20 // length of segment's ksuid
	nsLength            = byteLength - baseLength
	stringEncodedLength = 30
)

// IDs are 23 bytes:
// 	00-02: 3 ascii character namespace
//  03-06: uint32 BE UTC timestamp with custom epoch
//  07-22 byte: random "payload"
type ID struct {
	ns    Namespace
	ksuid ksuid.KSUID
}

// Constructs an ID from either a 23 byte ID representation OR a 20 byte KSUID
// representation.  If a 20 byte KSUID representation is used, the Unknown
// namespace prefix will be used
func FromBytes(b []byte) (id ID, err error) {
	switch len(b) {
	case byteLength:
		v, err := ksuid.FromBytes(b[nsLength:])
		if err != nil {
			return ID{}, err
		}

		copy(id.ns[:], b[0:nsLength])
		id.ksuid = v
		return id, nil

	default: // when b is a ksuid
		v, err := ksuid.FromBytes(b)
		if err != nil {
			return ID{}, err
		}
		id.ns = Unknown
		id.ksuid = v
		return id, nil
	}
}

// Parse a string representation of an ID
func Parse(s string) (id ID, err error) {
	if length := len(s); length == stringEncodedLength-nsLength {
		v, err := ksuid.Parse(s)
		if err != nil {
			return ID{}, err
		}
		return ID{
			ns:    Unknown,
			ksuid: v,
		}, nil
	} else if length != stringEncodedLength {
		return ID{}, errStringSize
	}

	ns, err := NewNamespace(s[0:nsLength])
	if err != nil {
		return ID{}, err
	}

	v, err := ksuid.Parse(s[nsLength:])
	if err != nil {
		return ID{}, err
	}

	id.ns = ns
	id.ksuid = v

	return id, nil
}

// constructs a new
func New(ns Namespace) ID {
	id, err := NewRandom(ns)
	if err != nil {
		panic(fmt.Sprintf("Couldn't generate KSUID, inconceivable! error: %v", err))
	}
	return id
}

// NewRandom using the ns provided and current time
func NewRandom(ns Namespace) (id ID, err error) {
	return NewRandomWithTime(ns, time.Now())
}

// NewRandomWithTime using the namespace and time provided
func NewRandomWithTime(ns Namespace, t time.Time) (id ID, err error) {
	v, err := ksuid.NewRandomWithTime(t)
	if err != nil {
		return ID{}, err
	}

	id.ns = ns
	id.ksuid = v

	return id, nil
}

// Append appends the string representation of i to b, returning a slice to a
// potentially larger memory area.
func (i ID) Append(b []byte) []byte {
	b = append(b, i.ns[:]...)
	return i.ksuid.Append(b)
}

// Bytes returns []byte representation of ID (23 bytes long)
func (i ID) Bytes() []byte {
	b := make([]byte, 0, byteLength)
	b = i.ns.Append(b)
	b = append(b, i.ksuid[:]...)
	return b
}

// Equal to provided ID
func (i ID) Equal(that ID) bool {
	return i.ns.Equal(that.ns) && bytes.Equal(i.ksuid[:], that.ksuid[:])
}

// IsNil returns true if this is a "nil" ID
func (i ID) IsNil() bool {
	return i.ns.Equal(Unknown) && i.ksuid.IsNil()
}

// MarshalJSON implements json.Marshaler
func (i ID) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, stringEncodedLength+2)
	b = append(b, '"')
	b = i.Append(b)
	b = append(b, '"')
	return b, nil
}

// KSUID returns underlying KSUID
func (i ID) KSUID() ksuid.KSUID {
	return i.ksuid
}

// Namespace returns underlying namespace
func (i ID) Namespace() Namespace {
	return i.ns
}

// String provides a string representation for the id
func (i ID) String() string {
	b := make([]byte, 0, stringEncodedLength)
	return string(i.Append(b))
}

// Time represents timestamp portion of the KSUID as a Time object
func (i ID) Time() time.Time {
	return i.ksuid.Time()
}

// UnmarshalJSON implements json.Unmarshaler
func (i *ID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*i = Nil
		return nil
	}

	id, err := Parse(s)
	if err != nil {
		return err
	}

	*i = id

	return nil
}

// Namespace to enable easy identification by user
type Namespace [nsLength]byte

// NewNamespace must be EXACTLY 3 ascii characters
func NewNamespace(v string) (Namespace, error) {
	b := []byte(v)
	if len(b) != nsLength {
		return Namespace{}, errNamespaceSize
	}

	var ns Namespace
	copy(ns[:], b[:])
	return ns, nil
}

// MustNamespace generates a namespace from the provided string or
// panics if the string is not 3 bytes
func MustNamespace(v string) Namespace {
	ns, err := NewNamespace(v)
	if err != nil {
		panic(fmt.Errorf("invalid namespace, %v: %w", v, err))
	}
	return ns
}

// Append the namespace to the provided byte array
func (n Namespace) Append(b []byte) []byte {
	return append(b, n[:]...)
}

func (n Namespace) raw() []byte {
	var b [3]byte = n
	return b[:]
}

// Bytes returns a []byte representation of the Namespace
func (n Namespace) Bytes() []byte {
	b := make([]byte, 0, nsLength)
	return n.Append(b)
}

// Equal returns true if the provided value is equal to the namespace
func (n Namespace) Equal(that interface{}) bool {
	switch v := that.(type) {
	case Namespace:
		return bytes.Equal(n.raw(), v.raw())
	case [3]byte:
		return bytes.Equal(n.raw(), v[:])
	case []byte:
		return bytes.Equal(n.raw(), v)
	case string:
		return bytes.Equal(n.raw(), []byte(v))
	default:
		return false
	}
}

// String view of Namespace
func (n Namespace) String() string {
	b := make([]byte, 0, nsLength)
	return string(n.Append(b))
}
