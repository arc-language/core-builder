// Package types defines the type system for the IR.
// This is platform-agnostic and represents abstract types.
package types

import (
	"fmt"
	"strings"
)

// TypeKind represents the category of a type
type TypeKind int

const (
	VoidKind TypeKind = iota
	IntegerKind
	FloatKind
	PointerKind
	ArrayKind
	StructKind
	FunctionKind
	VectorKind
	LabelKind
)

// Type is the interface all types implement
type Type interface {
	Kind() TypeKind
	String() string
	Equal(Type) bool
	// Size returns size in bits (0 for unsized types like void, label)
	BitSize() int
}

// VoidType represents the absence of a value
type VoidType struct{}

func (t *VoidType) Kind() TypeKind   { return VoidKind }
func (t *VoidType) String() string   { return "void" }
func (t *VoidType) BitSize() int     { return 0 }
func (t *VoidType) Equal(o Type) bool { return o.Kind() == VoidKind }

// IntType represents integers of arbitrary bit width
type IntType struct {
	BitWidth int
	Signed   bool
}

func (t *IntType) Kind() TypeKind { return IntegerKind }
func (t *IntType) String() string {
	prefix := "i"
	if !t.Signed {
		prefix = "u"
	}
	return fmt.Sprintf("%s%d", prefix, t.BitWidth)
}
func (t *IntType) BitSize() int { return t.BitWidth }
func (t *IntType) Equal(o Type) bool {
	if ot, ok := o.(*IntType); ok {
		return t.BitWidth == ot.BitWidth && t.Signed == ot.Signed
	}
	return false
}

// FloatType represents floating point types
type FloatType struct {
	BitWidth int // 16, 32, 64, 128
}

func (t *FloatType) Kind() TypeKind { return FloatKind }
func (t *FloatType) String() string {
	switch t.BitWidth {
	case 16:
		return "f16"
	case 32:
		return "f32"
	case 64:
		return "f64"
	case 128:
		return "f128"
	default:
		return fmt.Sprintf("f%d", t.BitWidth)
	}
}
func (t *FloatType) BitSize() int { return t.BitWidth }
func (t *FloatType) Equal(o Type) bool {
	if ot, ok := o.(*FloatType); ok {
		return t.BitWidth == ot.BitWidth
	}
	return false
}

// PointerType represents a pointer to another type
type PointerType struct {
	ElementType Type
	AddressSpace int // 0 = default, others for special memory regions
}

func (t *PointerType) Kind() TypeKind { return PointerKind }
func (t *PointerType) String() string {
	if t.AddressSpace != 0 {
		return fmt.Sprintf("ptr<%s, %d>", t.ElementType, t.AddressSpace)
	}
	return fmt.Sprintf("ptr<%s>", t.ElementType)
}
func (t *PointerType) BitSize() int { return 64 } // Platform-agnostic default
func (t *PointerType) Equal(o Type) bool {
	if ot, ok := o.(*PointerType); ok {
		return t.ElementType.Equal(ot.ElementType) && t.AddressSpace == ot.AddressSpace
	}
	return false
}

// ArrayType represents a fixed-size array
type ArrayType struct {
	ElementType Type
	Length      int64
}

func (t *ArrayType) Kind() TypeKind { return ArrayKind }
func (t *ArrayType) String() string {
	return fmt.Sprintf("[%d x %s]", t.Length, t.ElementType)
}
func (t *ArrayType) BitSize() int {
	return t.ElementType.BitSize() * int(t.Length)
}
func (t *ArrayType) Equal(o Type) bool {
	if ot, ok := o.(*ArrayType); ok {
		return t.Length == ot.Length && t.ElementType.Equal(ot.ElementType)
	}
	return false
}

// StructType represents a composite type
type StructType struct {
	Name     string
	Fields   []Type
	Packed   bool // If true, no padding between fields
}

func (t *StructType) Kind() TypeKind { return StructKind }
func (t *StructType) String() string {
	if t.Name != "" {
		return fmt.Sprintf("%%%s", t.Name)
	}
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}
	prefix := "{ "
	suffix := " }"
	if t.Packed {
		prefix = "<{ "
		suffix = " }>"
	}
	return prefix + strings.Join(fields, ", ") + suffix
}
func (t *StructType) BitSize() int {
	total := 0
	for _, f := range t.Fields {
		total += f.BitSize()
	}
	return total
}
func (t *StructType) Equal(o Type) bool {
	if ot, ok := o.(*StructType); ok {
		if t.Name != "" && ot.Name != "" {
			return t.Name == ot.Name
		}
		if len(t.Fields) != len(ot.Fields) {
			return false
		}
		for i := range t.Fields {
			if !t.Fields[i].Equal(ot.Fields[i]) {
				return false
			}
		}
		return t.Packed == ot.Packed
	}
	return false
}

// FunctionType represents a function signature
type FunctionType struct {
	ReturnType Type
	ParamTypes []Type
	Variadic   bool
}

func (t *FunctionType) Kind() TypeKind { return FunctionKind }
func (t *FunctionType) String() string {
	params := make([]string, len(t.ParamTypes))
	for i, p := range t.ParamTypes {
		params[i] = p.String()
	}
	if t.Variadic {
		params = append(params, "...")
	}
	return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), t.ReturnType)
}
func (t *FunctionType) BitSize() int { return 0 }
func (t *FunctionType) Equal(o Type) bool {
	if ot, ok := o.(*FunctionType); ok {
		if !t.ReturnType.Equal(ot.ReturnType) || t.Variadic != ot.Variadic {
			return false
		}
		if len(t.ParamTypes) != len(ot.ParamTypes) {
			return false
		}
		for i := range t.ParamTypes {
			if !t.ParamTypes[i].Equal(ot.ParamTypes[i]) {
				return false
			}
		}
		return true
	}
	return false
}

// VectorType represents SIMD vectors
type VectorType struct {
	ElementType Type
	Length      int
	Scalable    bool // For scalable vectors (SVE-like)
}

func (t *VectorType) Kind() TypeKind { return VectorKind }
func (t *VectorType) String() string {
	if t.Scalable {
		return fmt.Sprintf("<vscale x %d x %s>", t.Length, t.ElementType)
	}
	return fmt.Sprintf("<%d x %s>", t.Length, t.ElementType)
}
func (t *VectorType) BitSize() int {
	if t.Scalable {
		return 0 // Unknown at compile time
	}
	return t.ElementType.BitSize() * t.Length
}
func (t *VectorType) Equal(o Type) bool {
	if ot, ok := o.(*VectorType); ok {
		return t.Length == ot.Length && t.Scalable == ot.Scalable && t.ElementType.Equal(ot.ElementType)
	}
	return false
}

// LabelType represents a basic block label
type LabelType struct{}

func (t *LabelType) Kind() TypeKind   { return LabelKind }
func (t *LabelType) String() string   { return "label" }
func (t *LabelType) BitSize() int     { return 0 }
func (t *LabelType) Equal(o Type) bool { return o.Kind() == LabelKind }

// Common type constructors
var (
	Void  = &VoidType{}
	Label = &LabelType{}
	
	I1   = &IntType{BitWidth: 1, Signed: true}
	I8   = &IntType{BitWidth: 8, Signed: true}
	I16  = &IntType{BitWidth: 16, Signed: true}
	I32  = &IntType{BitWidth: 32, Signed: true}
	I64  = &IntType{BitWidth: 64, Signed: true}
	I128 = &IntType{BitWidth: 128, Signed: true}
	
	U8   = &IntType{BitWidth: 8, Signed: false}
	U16  = &IntType{BitWidth: 16, Signed: false}
	U32  = &IntType{BitWidth: 32, Signed: false}
	U64  = &IntType{BitWidth: 64, Signed: false}
	
	F16  = &FloatType{BitWidth: 16}
	F32  = &FloatType{BitWidth: 32}
	F64  = &FloatType{BitWidth: 64}
	F128 = &FloatType{BitWidth: 128}
)

// NewInt creates an integer type with the given bit width
func NewInt(bits int, signed bool) *IntType {
	return &IntType{BitWidth: bits, Signed: signed}
}

// NewFloat creates a float type with the given bit width
func NewFloat(bits int) *FloatType {
	return &FloatType{BitWidth: bits}
}

// NewPointer creates a pointer type
func NewPointer(elem Type) *PointerType {
	return &PointerType{ElementType: elem}
}

// NewPointerWithAddressSpace creates a pointer with address space
func NewPointerWithAddressSpace(elem Type, addrSpace int) *PointerType {
	return &PointerType{ElementType: elem, AddressSpace: addrSpace}
}

// NewArray creates an array type
func NewArray(elem Type, length int64) *ArrayType {
	return &ArrayType{ElementType: elem, Length: length}
}

// NewStruct creates a struct type
func NewStruct(name string, fields []Type, packed bool) *StructType {
	return &StructType{Name: name, Fields: fields, Packed: packed}
}

// NewFunction creates a function type
func NewFunction(ret Type, params []Type, variadic bool) *FunctionType {
	return &FunctionType{ReturnType: ret, ParamTypes: params, Variadic: variadic}
}

// NewVector creates a vector type
func NewVector(elem Type, length int) *VectorType {
	return &VectorType{ElementType: elem, Length: length}
}

// NewScalableVector creates a scalable vector type
func NewScalableVector(elem Type, minLength int) *VectorType {
	return &VectorType{ElementType: elem, Length: minLength, Scalable: true}
}

// IsInteger returns true if the type is an integer
func IsInteger(t Type) bool { return t.Kind() == IntegerKind }

// IsFloat returns true if the type is a float
func IsFloat(t Type) bool { return t.Kind() == FloatKind }

// IsPointer returns true if the type is a pointer
func IsPointer(t Type) bool { return t.Kind() == PointerKind }

// IsAggregate returns true if the type is an aggregate (struct or array)
func IsAggregate(t Type) bool {
	return t.Kind() == StructKind || t.Kind() == ArrayKind
}