// Package ir - instruction definitions
package ir

import (
	"fmt"
	"strings"

	"github.com/arc-language/core-builder/types"
)

// RetInst represents a return instruction
type RetInst struct {
	BaseInstruction
}

func (i *RetInst) String() string {
	if len(i.Ops) == 0 || i.Ops[0] == nil {
		return "ret void"
	}
	return fmt.Sprintf("ret %s %%%s", i.Ops[0].Type(), i.Ops[0].Name())
}

// BrInst represents an unconditional branch
type BrInst struct {
	BaseInstruction
	Target *BasicBlock
}

func (i *BrInst) String() string {
	return fmt.Sprintf("br label %%%s", i.Target.Name())
}

// CondBrInst represents a conditional branch
type CondBrInst struct {
	BaseInstruction
	Condition  Value
	TrueBlock  *BasicBlock
	FalseBlock *BasicBlock
}

func (i *CondBrInst) String() string {
	return fmt.Sprintf("br i1 %%%s, label %%%s, label %%%s",
		i.Condition.Name(), i.TrueBlock.Name(), i.FalseBlock.Name())
}

// SwitchInst represents a switch instruction
type SwitchInst struct {
	BaseInstruction
	Condition    Value
	DefaultBlock *BasicBlock
	Cases        []SwitchCase
}

type SwitchCase struct {
	Value *ConstantInt
	Block *BasicBlock
}

func (i *SwitchInst) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("switch %s %%%s, label %%%s [\n",
		i.Condition.Type(), i.Condition.Name(), i.DefaultBlock.Name()))
	for _, c := range i.Cases {
		sb.WriteString(fmt.Sprintf("    %s %d, label %%%s\n",
			c.Value.Type(), c.Value.Value, c.Block.Name()))
	}
	sb.WriteString("  ]")
	return sb.String()
}

// UnreachableInst represents unreachable code
type UnreachableInst struct {
	BaseInstruction
}

func (i *UnreachableInst) String() string {
	return "unreachable"
}

// BinaryInst represents binary operations
type BinaryInst struct {
	BaseInstruction
	NoSignedWrap   bool // nsw flag
	NoUnsignedWrap bool // nuw flag
	Exact          bool // exact flag (for div/shifts)
}

func (i *BinaryInst) String() string {
	flags := ""
	if i.NoUnsignedWrap {
		flags += " nuw"
	}
	if i.NoSignedWrap {
		flags += " nsw"
	}
	if i.Exact {
		flags += " exact"
	}

	lhs := i.Ops[0]
	rhs := i.Ops[1]

	return fmt.Sprintf("%%%s = %s%s %s %%%s, %%%s",
		i.ValName, i.Op, flags, lhs.Type(), lhs.Name(), rhs.Name())
}

// AllocaInst represents stack allocation
type AllocaInst struct {
	BaseInstruction
	AllocatedType types.Type
	NumElements   Value // nil means 1 element
	Alignment     int
}

func (i *AllocaInst) String() string {
	result := fmt.Sprintf("%%%s = alloca %s", i.ValName, i.AllocatedType)
	if i.NumElements != nil {
		result += fmt.Sprintf(", %s %%%s", i.NumElements.Type(), i.NumElements.Name())
	}
	if i.Alignment > 0 {
		result += fmt.Sprintf(", align %d", i.Alignment)
	}
	return result
}

// LoadInst represents a memory load
type LoadInst struct {
	BaseInstruction
	Volatile  bool
	Alignment int
}

func (i *LoadInst) String() string {
	vol := ""
	if i.Volatile {
		vol = "volatile "
	}
	ptr := i.Ops[0]
	result := fmt.Sprintf("%%%s = load %s%s, %s %%%s",
		i.ValName, vol, i.ValType, ptr.Type(), ptr.Name())
	if i.Alignment > 0 {
		result += fmt.Sprintf(", align %d", i.Alignment)
	}
	return result
}

// StoreInst represents a memory store
type StoreInst struct {
	BaseInstruction
	Volatile  bool
	Alignment int
}

func (i *StoreInst) String() string {
	vol := ""
	if i.Volatile {
		vol = "volatile "
	}
	val := i.Ops[0]
	ptr := i.Ops[1]
	result := fmt.Sprintf("store %s%s %%%s, %s %%%s",
		vol, val.Type(), val.Name(), ptr.Type(), ptr.Name())
	if i.Alignment > 0 {
		result += fmt.Sprintf(", align %d", i.Alignment)
	}
	return result
}

// GetElementPtrInst represents pointer arithmetic
type GetElementPtrInst struct {
	BaseInstruction
	SourceElementType types.Type
	InBounds          bool
}

func (i *GetElementPtrInst) String() string {
	inbounds := ""
	if i.InBounds {
		inbounds = "inbounds "
	}
	ptr := i.Ops[0]
	
	var indices []string
	for _, idx := range i.Ops[1:] {
		indices = append(indices, fmt.Sprintf("%s %%%s", idx.Type(), idx.Name()))
	}
	
	return fmt.Sprintf("%%%s = getelementptr %s%s, %s %%%s, %s",
		i.ValName, inbounds, i.SourceElementType, ptr.Type(), ptr.Name(),
		strings.Join(indices, ", "))
}

// CastInst represents type conversion operations
type CastInst struct {
	BaseInstruction
	DestType types.Type
}

func (i *CastInst) String() string {
	src := i.Ops[0]
	return fmt.Sprintf("%%%s = %s %s %%%s to %s",
		i.ValName, i.Op, src.Type(), src.Name(), i.DestType)
}

// ICmpInst represents integer comparison
type ICmpInst struct {
	BaseInstruction
	Predicate ICmpPredicate
}

func (i *ICmpInst) String() string {
	lhs := i.Ops[0]
	rhs := i.Ops[1]
	return fmt.Sprintf("%%%s = icmp %s %s %%%s, %%%s",
		i.ValName, i.Predicate, lhs.Type(), lhs.Name(), rhs.Name())
}

// FCmpInst represents floating point comparison
type FCmpInst struct {
	BaseInstruction
	Predicate FCmpPredicate
}

func (i *FCmpInst) String() string {
	lhs := i.Ops[0]
	rhs := i.Ops[1]
	return fmt.Sprintf("%%%s = fcmp %s %s %%%s, %%%s",
		i.ValName, i.Predicate, lhs.Type(), lhs.Name(), rhs.Name())
}

// PhiInst represents a phi node
type PhiInst struct {
	BaseInstruction
	Incoming []PhiIncoming
}

type PhiIncoming struct {
	Value Value
	Block *BasicBlock
}

func (i *PhiInst) String() string {
	var incomings []string
	for _, inc := range i.Incoming {
		incomings = append(incomings,
			fmt.Sprintf("[ %%%s, %%%s ]", inc.Value.Name(), inc.Block.Name()))
	}
	return fmt.Sprintf("%%%s = phi %s %s",
		i.ValName, i.ValType, strings.Join(incomings, ", "))
}

func (i *PhiInst) AddIncoming(v Value, b *BasicBlock) {
	i.Incoming = append(i.Incoming, PhiIncoming{Value: v, Block: b})
}

// SelectInst represents a select (ternary) operation
type SelectInst struct {
	BaseInstruction
}

func (i *SelectInst) String() string {
	cond := i.Ops[0]
	trueVal := i.Ops[1]
	falseVal := i.Ops[2]
	return fmt.Sprintf("%%%s = select i1 %%%s, %s %%%s, %s %%%s",
		i.ValName, cond.Name(),
		trueVal.Type(), trueVal.Name(),
		falseVal.Type(), falseVal.Name())
}

// CallInst represents a function call
type CallInst struct {
	BaseInstruction
	Callee     *Function
	CalleeName string // For indirect calls or declarations
	IsTailCall bool
}

func (i *CallInst) String() string {
	tail := ""
	if i.IsTailCall {
		tail = "tail "
	}
	
	var args []string
	for _, arg := range i.Ops {
		if arg != nil {
			args = append(args, fmt.Sprintf("%s %%%s", arg.Type(), arg.Name()))
		}
	}
	
	calleeName := i.CalleeName
	if i.Callee != nil {
		calleeName = i.Callee.Name()
	}
	
	// FIX: Check for nil ValType to prevent panic
	if i.ValType == nil || i.ValType.Kind() == types.VoidKind {
		return fmt.Sprintf("%scall void @%s(%s)", tail, calleeName, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%%%s = %scall %s @%s(%s)",
		i.ValName, tail, i.ValType, calleeName, strings.Join(args, ", "))
}

// ExtractValueInst extracts a value from an aggregate
type ExtractValueInst struct {
	BaseInstruction
	Indices []int
}

func (i *ExtractValueInst) String() string {
	agg := i.Ops[0]
	indices := make([]string, len(i.Indices))
	for j, idx := range i.Indices {
		indices[j] = fmt.Sprintf("%d", idx)
	}
	return fmt.Sprintf("%%%s = extractvalue %s %%%s, %s",
		i.ValName, agg.Type(), agg.Name(), strings.Join(indices, ", "))
}

// InsertValueInst inserts a value into an aggregate
type InsertValueInst struct {
	BaseInstruction
	Indices []int
}

func (i *InsertValueInst) String() string {
	agg := i.Ops[0]
	val := i.Ops[1]
	indices := make([]string, len(i.Indices))
	for j, idx := range i.Indices {
		indices[j] = fmt.Sprintf("%d", idx)
	}
	return fmt.Sprintf("%%%s = insertvalue %s %%%s, %s %%%s, %s",
		i.ValName, agg.Type(), agg.Name(),
		val.Type(), val.Name(),
		strings.Join(indices, ", "))
}