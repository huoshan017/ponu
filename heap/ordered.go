package heap

import "golang.org/x/exp/constraints"

type Ordered interface {
	Greater(Ordered) bool
	GreaterEqual(Ordered) bool
	Less(Ordered) bool
	LessEqual(Ordered) bool
}

type Int8 int8

func (i Int8) Greater(i2 Ordered) bool {
	return i > i2.(Int8)
}

func (i Int8) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Int8)
}

func (i Int8) Less(i2 Ordered) bool {
	return i < i2.(Int8)
}

func (i Int8) LessEqual(i2 Ordered) bool {
	return i <= i2.(Int8)
}

type Int16 int16

func (i Int16) Greater(i2 Ordered) bool {
	return i > i2.(Int16)
}

func (i Int16) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Int16)
}

func (i Int16) Less(i2 Ordered) bool {
	return i < i2.(Int16)
}

func (i Int16) LessEqual(i2 Ordered) bool {
	return i <= i2.(Int16)
}

type Int32 int32

func (i Int32) Greater(i2 Ordered) bool {
	return i > i2.(Int32)
}

func (i Int32) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Int32)
}

func (i Int32) Less(i2 Ordered) bool {
	return i < i2.(Int32)
}

func (i Int32) LessEqual(i2 Ordered) bool {
	return i <= i2.(Int32)
}

type Int64 int64

func (i Int64) Greater(i2 Ordered) bool {
	return i > i2.(Int64)
}

func (i Int64) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Int64)
}

func (i Int64) Less(i2 Ordered) bool {
	return i < i2.(Int64)
}

func (i Int64) LessEqual(i2 Ordered) bool {
	return i <= i2.(Int64)
}

type Int int

func (i Int) Greater(i2 Ordered) bool {
	return i > i2.(Int)
}

func (i Int) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Int)
}

func (i Int) Less(i2 Ordered) bool {
	return i < i2.(Int)
}

func (i Int) LessEqual(i2 Ordered) bool {
	return i <= i2.(Int)
}

type Uint8 uint8

func (i Uint8) Greater(i2 Ordered) bool {
	return i > i2.(Uint8)
}

func (i Uint8) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Uint8)
}

func (i Uint8) Less(i2 Ordered) bool {
	return i < i2.(Uint8)
}

func (i Uint8) LessEqual(i2 Ordered) bool {
	return i <= i2.(Uint8)
}

type Uint16 uint16

func (i Uint16) Greater(i2 Ordered) bool {
	return i > i2.(Uint16)
}

func (i Uint16) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Uint16)
}

func (i Uint16) Less(i2 Ordered) bool {
	return i < i2.(Uint16)
}

func (i Uint16) LessEqual(i2 Ordered) bool {
	return i <= i2.(Uint16)
}

type Uint32 int32

func (i Uint32) Greater(i2 Ordered) bool {
	return i > i2.(Uint32)
}

func (i Uint32) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Uint32)
}

func (i Uint32) Less(i2 Ordered) bool {
	return i < i2.(Uint32)
}

func (i Uint32) LessEqual(i2 Ordered) bool {
	return i <= i2.(Uint32)
}

type Uint64 uint64

func (i Uint64) Greater(i2 Ordered) bool {
	return i > i2.(Uint64)
}

func (i Uint64) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Uint64)
}

func (i Uint64) Less(i2 Ordered) bool {
	return i < i2.(Uint64)
}

func (i Uint64) LessEqual(i2 Ordered) bool {
	return i <= i2.(Uint64)
}

type Uint uint

func (i Uint) Greater(i2 Ordered) bool {
	return i > i2.(Uint)
}

func (i Uint) GreaterEqual(i2 Ordered) bool {
	return i >= i2.(Uint)
}

func (i Uint) Less(i2 Ordered) bool {
	return i < i2.(Uint)
}

func (i Uint) LessEqual(i2 Ordered) bool {
	return i <= i2.(Uint)
}

type Float32 float32

func (f Float32) Greater(f2 Ordered) bool {
	return f > f2.(Float32)
}

func (f Float32) GreaterEqual(f2 Ordered) bool {
	return f >= f2.(Float32)
}

func (f Float32) Less(f2 Ordered) bool {
	return f < f2.(Float32)
}

func (f Float32) LessEqual(f2 Ordered) bool {
	return f <= f2.(Float32)
}

type Float64 float64

func (f Float64) Greater(f2 Ordered) bool {
	return f > f2.(Float64)
}

func (f Float64) GreaterEqual(f2 Ordered) bool {
	return f >= f2.(Float64)
}

func (f Float64) Less(f2 Ordered) bool {
	return f < f2.(Float64)
}

func (f Float64) LessEqual(f2 Ordered) bool {
	return f <= f2.(Float64)
}

type MyOrdered[T constraints.Ordered] struct {
	value T
}

func (o MyOrdered[T]) Greater(o2 Ordered) bool {
	return o.value > o2.(MyOrdered[T]).value
}

func (o MyOrdered[T]) GreaterEqual(o2 Ordered) bool {
	return o.value >= o2.(MyOrdered[T]).value
}

func (o MyOrdered[T]) Less(o2 Ordered) bool {
	return o.value < o2.(MyOrdered[T]).value
}

func (o MyOrdered[T]) LessEqual(o2 Ordered) bool {
	return o.value <= o2.(MyOrdered[T]).value
}
