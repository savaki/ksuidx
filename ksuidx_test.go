package ksuidx

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/segmentio/ksuid"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewRandomWithTime(t *testing.T) {
	ns := MustNamespace("bla")
	v := New(ns)
	if got, want := v.String(), ns.String(); !strings.HasPrefix(got, want) {
		t.Fatalf("got %v; want prefix %v", got, want)
	}
}

func TestFromBytes(t *testing.T) {
	t.Run("from ksuid", func(t *testing.T) {
		v := ksuid.New()
		id, err := FromBytes(v.Bytes())
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if got, want := id.Namespace().String(), Unknown.String(); got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("from id", func(t *testing.T) {
		ns := MustNamespace("bla")
		want := New(ns)

		got, err := FromBytes(want.Bytes())
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})
}

func TestID_KSUID(t *testing.T) {
	ns := MustNamespace("bla")
	v := New(ns)
	if got, want := v.ksuid, v.KSUID(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestNewNamespace(t *testing.T) {
	t.Run("bad", func(t *testing.T) {
		_, err := NewNamespace("too-long")
		if got, want := err, errNamespaceSize; got != want {
			t.Fatalf("got %v; want %v", got, want)
		}
	})
}

func TestFromString(t *testing.T) {
	ns := MustNamespace("bla")
	want := New(ns)
	got, err := Parse(want.String())
	if err != nil {
		t.Fatalf("got %v; want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v; want %v", got, want)
	}
}

func TestNamespace_Equal(t *testing.T) {
	ns := MustNamespace("bla")
	testCases := map[string]struct {
		Input interface{}
		Want  bool
	}{
		"namespace": {
			Input: ns,
			Want:  true,
		},
		"namespace - no match": {
			Input: MustNamespace("eek"),
			Want:  false,
		},
		"[3]byte": {
			Input: [3]byte{'a', 'b', 'c'},
			Want:  false,
		},
		"[]byte": {
			Input: ns.Bytes(),
			Want:  true,
		},
		"string": {
			Input: ns.String(),
			Want:  true,
		},
		"nope": {
			Input: 123,
			Want:  false,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := ns.Equal(tc.Input)
			if want := tc.Want; got != want {
				t.Fatalf("got %v; want %v", got, want)
			}
		})
	}
}

func TestNamespace_Bytes(t *testing.T) {
	raw := "bla"
	ns := MustNamespace(raw)
	if got, want := ns.Bytes(), []byte(raw); !bytes.Equal(got, want) {
		t.Fatalf("got %v; want %v", string(got), string(want))
	}
}

func TestID_Time(t *testing.T) {
	ns := MustNamespace("bla")
	id := New(ns)
	got := time.Now().Unix() - id.Time().Unix()
	if got > 1 {
		t.Fatalf("got %v; want <= 1", got)
	}
}

func TestID_UnmarshalJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ns := MustNamespace("bla")
		want := New(ns)

		b, err := want.MarshalJSON()
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		var got ID
		err = json.Unmarshal(b, &got)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
		if !want.Equal(got) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("null", func(t *testing.T) {
		var got ID
		err := json.Unmarshal([]byte(`null`), &got)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}
	})

	t.Run("bad json", func(t *testing.T) {
		var got ID
		err := (&got).UnmarshalJSON([]byte(`"bad-json`))
		var want *json.SyntaxError
		ok := errors.As(err, &want)
		if !ok {
			t.Fatalf("got %v; want true", ok)
		}
	})

	t.Run("bad ksuid", func(t *testing.T) {
		var got ID
		err := json.Unmarshal([]byte(`"invalid-ksuid"`), &got)
		if got, want := err, errStringSize; got != want {
			t.Fatalf("got %v; want %v", got, err)
		}
	})
}

func TestID_IsNil(t *testing.T) {
	ok := Nil.IsNil()
	if !ok {
		t.Fatalf("got %v; want true", ok)
	}
}
