// Package ir defines the intermediate representation nodes.
// This is the core data structure for the IR tree.
package ir

import (
	"fmt"
	"strings"

	"github.com/arc-language/core-builder/types"
)

// Value is the interface for all IR values
type Value interface {
	Type() types.Type
	Name() string
	SetName(string)
	String() string
}

// User is a Value that references other Values
type User interface {
	Value
	Operands() []Value
	SetOperand(int, Value)
	NumOperands() int
}

// Instruction is an operation that produces a value
type Instruction interface {
	User
	Opcode() Opcode
	Parent() *BasicBlock
	SetParent(*BasicBlock)
	IsTerminator() bool
}

// Opcode represents the operation type
type Opcode int

const (
	// Terminator instructions
	OpRet Opcode = iota
	OpBr
	OpCondBr
	OpSwitch
	OpUnreachable

	// Binary operations
	OpAdd
	OpSub
	OpMul
	OpUDiv
	OpSDiv
	OpURem
	OpSRem
	OpFAdd
	OpFSub
	OpFMul
	OpFDiv
	OpFRem

	// Bitwise binary operations
	OpShl
	OpLShr
	OpAShr
	OpAnd
	OpOr
	OpXor

	// Memory operations
	OpAlloca
	OpLoad
	OpStore
	OpGetElementPtr

	// Cast operations
	OpTrunc
	OpZExt
	OpSExt
	OpFPTrunc
	OpFPExt
	OpFPToUI
	OpFPToSI
	OpUIToFP
	OpSIToFP
	OpPtrToInt
	OpIntToPtr
	OpBitcast

	// Other operations
	OpICmp
	OpFCmp
	OpPhi
	OpSelect
	OpCall
	OpExtractValue
	OpInsertValue
)

var opcodeNames = map[Opcode]string{
	OpRet:         "ret",
	OpBr:          "br",
	OpCondBr:      "br",
	OpSwitch:      "switch",
	OpUnreachable: "unreachable",
	OpAdd:         "add",
	OpSub:         "sub",
	OpMul:         "mul",
	OpUDiv:        "udiv",
	OpSDiv:        "sdiv",
	OpURem:        "urem",
	OpSRem:        "srem",
	OpFAdd:        "fadd",
	OpFSub:        "fsub",
	OpFMul:        "fmul",
	OpFDiv:        "fdiv",
	OpFRem:        "frem",
	OpShl:         "shl",
	OpLShr:        "lshr",
	OpAShr:        "ashr",
	OpAnd:         "and",
	OpOr:          "or",
	OpXor:         "xor",
	OpAlloca:      "alloca",
	OpLoad:        "load",
	OpStore:       "store",
	OpGetElementPtr: "getelementptr",
	OpTrunc:       "trunc",
	OpZExt:        "zext",
	OpSExt:        "sext",
	OpFPTrunc:     "fptrunc",
	OpFPExt:       "fpext",
	OpFPToUI:      "fptoui",
	OpFPToSI:      "fptosi",
	OpUIToFP:      "uitofp",
	OpSIToFP:      "sitofp",
	OpPtrToInt:    "ptrtoint",
	OpIntToPtr:    "inttoptr",
	OpBitcast:     "bitcast",
	OpICmp:        "icmp",
	OpFCmp:        "fcmp",
	OpPhi:         "phi",
	OpSelect:      "select",
	OpCall:        "call",
	OpExtractValue: "extractvalue",
	OpInsertValue: "insertvalue",
}

func (op Opcode) String() string {
	if name, ok := opcodeNames[op]; ok {
		return name
	}
	return fmt.Sprintf("op%d", op)
}

// ICmpPredicate represents integer comparison predicates
type ICmpPredicate int

const (
	ICmpEQ  ICmpPredicate = iota // equal
	ICmpNE                       // not equal
	ICmpUGT                      // unsigned greater than
	ICmpUGE                      // unsigned greater or equal
	ICmpULT                      // unsigned less than
	ICmpULE                      // unsigned less or equal
	ICmpSGT                      // signed greater than
	ICmpSGE                      // signed greater or equal
	ICmpSLT                      // signed less than
	ICmpSLE                      // signed less or equal
)

var icmpNames = map[ICmpPredicate]string{
	ICmpEQ: "eq", ICmpNE: "ne",
	ICmpUGT: "ugt", ICmpUGE: "uge", ICmpULT: "ult", ICmpULE: "ule",
	ICmpSGT: "sgt", ICmpSGE: "sge", ICmpSLT: "slt", ICmpSLE: "sle",
}

func (p ICmpPredicate) String() string { return icmpNames[p] }

// FCmpPredicate represents floating point comparison predicates
type FCmpPredicate int

const (
	FCmpFalse FCmpPredicate = iota // always false
	FCmpOEQ                        // ordered and equal
	FCmpOGT                        // ordered and greater than
	FCmpOGE                        // ordered and greater or equal
	FCmpOLT                        // ordered and less than
	FCmpOLE                        // ordered and less or equal
	FCmpONE                        // ordered and not equal
	FCmpORD                        // ordered (no NaN)
	FCmpUNO                        // unordered (either NaN)
	FCmpUEQ                        // unordered or equal
	FCmpUGT                        // unordered or greater than
	FCmpUGE                        // unordered or greater or equal
	FCmpULT                        // unordered or less than
	FCmpULE                        // unordered or less or equal
	FCmpUNE                        // unordered or not equal
	FCmpTrue                       // always true
)

var fcmpNames = map[FCmpPredicate]string{
	FCmpFalse: "false", FCmpOEQ: "oeq", FCmpOGT: "ogt", FCmpOGE: "oge",
	FCmpOLT: "olt", FCmpOLE: "ole", FCmpONE: "one", FCmpORD: "ord",
	FCmpUNO: "uno", FCmpUEQ: "ueq", FCmpUGT: "ugt", FCmpUGE: "uge",
	FCmpULT: "ult", FCmpULE: "ule", FCmpUNE: "une", FCmpTrue: "true",
}

func (p FCmpPredicate) String() string { return fcmpNames[p] }

// BaseValue provides common functionality for all values
type BaseValue struct {
	ValName string
	ValType types.Type
}

func (v *BaseValue) Name() string       { return v.ValName }
func (v *BaseValue) SetName(n string)   { v.ValName = n }
func (v *BaseValue) Type() types.Type   { return v.ValType }
func (v *BaseValue) SetType(t types.Type) { v.ValType = t }

// BaseInstruction provides common functionality for instructions
type BaseInstruction struct {
	BaseValue
	Ops    []Value
	Parent_ *BasicBlock
	Op     Opcode
}

func (i *BaseInstruction) Opcode() Opcode           { return i.Op }
func (i *BaseInstruction) Parent() *BasicBlock      { return i.Parent_ }
func (i *BaseInstruction) SetParent(b *BasicBlock)  { i.Parent_ = b }
func (i *BaseInstruction) Operands() []Value        { return i.Ops }
func (i *BaseInstruction) NumOperands() int         { return len(i.Ops) }
func (i *BaseInstruction) SetOperand(idx int, v Value) {
	// Grow slice if needed
	for len(i.Ops) <= idx {
		i.Ops = append(i.Ops, nil)
	}
	i.Ops[idx] = v
}
func (i *BaseInstruction) IsTerminator() bool {
	switch i.Op {
	case OpRet, OpBr, OpCondBr, OpSwitch, OpUnreachable:
		return true
	}
	return false
}

// Constant values
type Constant interface {
	Value
	isConstant()
}

// ConstantInt represents an integer constant
type ConstantInt struct {
	BaseValue
	Value int64
}

func (c *ConstantInt) isConstant() {}
func (c *ConstantInt) String() string {
	return fmt.Sprintf("%s %d", c.ValType, c.Value)
}

// ConstantFloat represents a floating point constant
type ConstantFloat struct {
	BaseValue
	Value float64
}

func (c *ConstantFloat) isConstant() {}
func (c *ConstantFloat) String() string {
	return fmt.Sprintf("%s %g", c.ValType, c.Value)
}

// ConstantNull represents a null pointer
type ConstantNull struct {
	BaseValue
}

func (c *ConstantNull) isConstant() {}
func (c *ConstantNull) String() string {
	return fmt.Sprintf("%s null", c.ValType)
}

// ConstantUndef represents an undefined value
type ConstantUndef struct {
	BaseValue
}

func (c *ConstantUndef) isConstant() {}
func (c *ConstantUndef) String() string {
	return fmt.Sprintf("%s undef", c.ValType)
}

// ConstantArray represents an array constant
type ConstantArray struct {
	BaseValue
	Elements []Constant
}

func (c *ConstantArray) isConstant() {}
func (c *ConstantArray) String() string {
	elems := make([]string, len(c.Elements))
	for i, e := range c.Elements {
		elems[i] = e.String()
	}
	return fmt.Sprintf("%s [%s]", c.ValType, strings.Join(elems, ", "))
}

// ConstantStruct represents a struct constant
type ConstantStruct struct {
	BaseValue
	Fields []Constant
}

func (c *ConstantStruct) isConstant() {}
func (c *ConstantStruct) String() string {
	fields := make([]string, len(c.Fields))
	for i, f := range c.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("%s { %s }", c.ValType, strings.Join(fields, ", "))
}

// ConstantZero represents a zero initializer for any type
type ConstantZero struct {
	BaseValue
}

func (c *ConstantZero) isConstant() {}
func (c *ConstantZero) String() string {
	return fmt.Sprintf("%s zeroinitializer", c.ValType)
}

// Global represents a global variable or constant
type Global struct {
	BaseValue
	Initializer Constant
	IsConstant  bool
	Linkage     Linkage
	AddressSpace int
}

func (g *Global) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("@%s =", g.ValName))
	parts = append(parts, g.Linkage.String())
	if g.IsConstant {
		parts = append(parts, "constant")
	} else {
		parts = append(parts, "global")
	}
	if g.Initializer != nil {
		parts = append(parts, g.Initializer.String())
	} else {
		parts = append(parts, g.ValType.String())
	}
	return strings.Join(parts, " ")
}

// Linkage types for globals and functions
type Linkage int

const (
	ExternalLinkage Linkage = iota
	InternalLinkage
	PrivateLinkage
	LinkOnceODRLinkage
	WeakODRLinkage
	CommonLinkage
)

func (l Linkage) String() string {
	switch l {
	case ExternalLinkage:
		return "external"
	case InternalLinkage:
		return "internal"
	case PrivateLinkage:
		return "private"
	case LinkOnceODRLinkage:
		return "linkonce_odr"
	case WeakODRLinkage:
		return "weak_odr"
	case CommonLinkage:
		return "common"
	}
	return "external"
}

// Argument represents a function argument
type Argument struct {
	BaseValue
	Index  int
	Parent *Function
}

func (a *Argument) String() string {
	if a.ValName != "" {
		return fmt.Sprintf("%s %%%s", a.ValType, a.ValName)
	}
	return fmt.Sprintf("%s %%%d", a.ValType, a.Index)
}

// BasicBlock represents a basic block in the CFG
type BasicBlock struct {
	BaseValue
	Instructions []Instruction
	Parent       *Function
	Predecessors []*BasicBlock
	Successors   []*BasicBlock
}

func NewBasicBlock(name string) *BasicBlock {
	return &BasicBlock{
		BaseValue: BaseValue{ValName: name, ValType: types.Label},
	}
}

func (b *BasicBlock) String() string {
	var sb strings.Builder
	sb.WriteString(b.ValName)
	sb.WriteString(":\n")
	for _, inst := range b.Instructions {
		sb.WriteString("  ")
		sb.WriteString(inst.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *BasicBlock) AddInstruction(inst Instruction) {
	inst.SetParent(b)
	b.Instructions = append(b.Instructions, inst)
}

func (b *BasicBlock) Terminator() Instruction {
	if len(b.Instructions) == 0 {
		return nil
	}
	last := b.Instructions[len(b.Instructions)-1]
	if last.IsTerminator() {
		return last
	}
	return nil
}

// Function represents a function
type Function struct {
	BaseValue
	FuncType   *types.FunctionType
	Blocks     []*BasicBlock
	Arguments  []*Argument
	Linkage    Linkage
	Parent     *Module
	Attributes []FuncAttribute
}

type FuncAttribute int

const (
	AttrNoReturn FuncAttribute = iota
	AttrNoUnwind
	AttrReadOnly
	AttrReadNone
	AttrAlwaysInline
	AttrNoInline
)

func NewFunction(name string, fnType *types.FunctionType) *Function {
	f := &Function{
		BaseValue: BaseValue{ValName: name, ValType: fnType},
		FuncType:  fnType,
	}
	// Create arguments
	for i, paramType := range fnType.ParamTypes {
		arg := &Argument{
			BaseValue: BaseValue{ValType: paramType},
			Index:     i,
			Parent:    f,
		}
		f.Arguments = append(f.Arguments, arg)
	}
	return f
}

func (f *Function) AddBlock(b *BasicBlock) {
	b.Parent = f
	f.Blocks = append(f.Blocks, b)
}

func (f *Function) EntryBlock() *BasicBlock {
	if len(f.Blocks) > 0 {
		return f.Blocks[0]
	}
	return nil
}

func (f *Function) String() string {
	var sb strings.Builder
	
	// Declaration vs definition
	if len(f.Blocks) == 0 {
		sb.WriteString("declare ")
	} else {
		sb.WriteString("define ")
	}
	
	sb.WriteString(f.Linkage.String())
	sb.WriteString(" ")
	sb.WriteString(f.FuncType.ReturnType.String())
	sb.WriteString(" @")
	sb.WriteString(f.ValName)
	sb.WriteString("(")
	
	args := make([]string, len(f.Arguments))
	for i, arg := range f.Arguments {
		if arg.ValName != "" {
			args[i] = fmt.Sprintf("%s %%%s", arg.ValType, arg.ValName)
		} else {
			args[i] = arg.ValType.String()
		}
	}
	sb.WriteString(strings.Join(args, ", "))
	if f.FuncType.Variadic {
		if len(args) > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("...")
	}
	sb.WriteString(")")
	
	// Attributes
	for _, attr := range f.Attributes {
		sb.WriteString(" ")
		switch attr {
		case AttrNoReturn:
			sb.WriteString("noreturn")
		case AttrNoUnwind:
			sb.WriteString("nounwind")
		case AttrReadOnly:
			sb.WriteString("readonly")
		case AttrReadNone:
			sb.WriteString("readnone")
		case AttrAlwaysInline:
			sb.WriteString("alwaysinline")
		case AttrNoInline:
			sb.WriteString("noinline")
		}
	}
	
	if len(f.Blocks) > 0 {
		sb.WriteString(" {\n")
		for _, block := range f.Blocks {
			sb.WriteString(block.String())
		}
		sb.WriteString("}")
	}
	
	return sb.String()
}

// Module represents a compilation unit
type Module struct {
	Name      string
	Functions []*Function
	Globals   []*Global
	Types     map[string]*types.StructType
	DataLayout string
	TargetTriple string
}

func NewModule(name string) *Module {
	return &Module{
		Name:  name,
		Types: make(map[string]*types.StructType),
	}
}

func (m *Module) AddFunction(f *Function) {
	f.Parent = m
	m.Functions = append(m.Functions, f)
}

func (m *Module) AddGlobal(g *Global) {
	m.Globals = append(m.Globals, g)
}

func (m *Module) GetFunction(name string) *Function {
	for _, f := range m.Functions {
		if f.ValName == name {
			return f
		}
	}
	return nil
}

func (m *Module) GetGlobal(name string) *Global {
	for _, g := range m.Globals {
		if g.ValName == name {
			return g
		}
	}
	return nil
}

func (m *Module) String() string {
	var sb strings.Builder
	
	if m.DataLayout != "" {
		sb.WriteString(fmt.Sprintf("target datalayout = \"%s\"\n", m.DataLayout))
	}
	if m.TargetTriple != "" {
		sb.WriteString(fmt.Sprintf("target triple = \"%s\"\n", m.TargetTriple))
	}
	if m.DataLayout != "" || m.TargetTriple != "" {
		sb.WriteString("\n")
	}
	
	// Named types
	for name, typ := range m.Types {
		sb.WriteString(fmt.Sprintf("%%%s = type %s\n", name, typ.String()))
	}
	if len(m.Types) > 0 {
		sb.WriteString("\n")
	}
	
	// Globals
	for _, g := range m.Globals {
		sb.WriteString(g.String())
		sb.WriteString("\n")
	}
	if len(m.Globals) > 0 {
		sb.WriteString("\n")
	}
	
	// Functions
	for i, f := range m.Functions {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.String())
		sb.WriteString("\n")
	}
	
	return sb.String()
}