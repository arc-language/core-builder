// Package builder provides the IR Builder pattern for constructing IR.
// This is the main API users interact with to create IR.
package builder

import (
	"fmt"

	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
)

// Builder constructs IR instructions
type Builder struct {
	module       *ir.Module
	currentFunc  *ir.Function
	currentBlock *ir.BasicBlock
	insertPoint  int // -1 means append to end
	nameCounter  int
}

// New creates a new IR builder
func New() *Builder {
	return &Builder{
		insertPoint: -1,
	}
}

// NewWithModule creates a builder with an existing module
func NewWithModule(m *ir.Module) *Builder {
	return &Builder{
		module:      m,
		insertPoint: -1,
	}
}

// Module returns the current module
func (b *Builder) Module() *ir.Module {
	return b.module
}

// CreateModule creates a new module
func (b *Builder) CreateModule(name string) *ir.Module {
	b.module = ir.NewModule(name)
	return b.module
}

// CurrentFunction returns the function being built
func (b *Builder) CurrentFunction() *ir.Function {
	return b.currentFunc
}

// CurrentBlock returns the current basic block
func (b *Builder) CurrentBlock() *ir.BasicBlock {
	return b.currentBlock
}

// SetInsertPoint sets where instructions will be inserted
func (b *Builder) SetInsertPoint(block *ir.BasicBlock) {
	b.currentBlock = block
	b.currentFunc = block.Parent
	b.insertPoint = -1
}

// SetInsertPointBefore sets insertion point before an instruction
func (b *Builder) SetInsertPointBefore(inst ir.Instruction) {
	b.currentBlock = inst.Parent()
	for i, in := range b.currentBlock.Instructions {
		if in == inst {
			b.insertPoint = i
			return
		}
	}
}

// GetInsertBlock returns the current insertion block
func (b *Builder) GetInsertBlock() *ir.BasicBlock {
	return b.currentBlock
}

// generateName creates a unique name for unnamed values
func (b *Builder) generateName() string {
	name := fmt.Sprintf("%d", b.nameCounter)
	b.nameCounter++
	return name
}

// insert adds an instruction at the current insertion point
func (b *Builder) insert(inst ir.Instruction) {
	if b.currentBlock == nil {
		panic("no insertion block set")
	}
	if b.insertPoint < 0 {
		b.currentBlock.AddInstruction(inst)
	} else {
		// Insert at specific position
		insts := b.currentBlock.Instructions
		newInsts := make([]ir.Instruction, len(insts)+1)
		copy(newInsts, insts[:b.insertPoint])
		newInsts[b.insertPoint] = inst
		copy(newInsts[b.insertPoint+1:], insts[b.insertPoint:])
		b.currentBlock.Instructions = newInsts
		inst.SetParent(b.currentBlock)
		b.insertPoint++
	}
}

// ============================================================================
// Module-level operations
// ============================================================================

// CreateFunction creates a new function in the module
func (b *Builder) CreateFunction(name string, retType types.Type, params []types.Type, variadic bool) *ir.Function {
	fnType := types.NewFunction(retType, params, variadic)
	fn := ir.NewFunction(name, fnType)
	if b.module != nil {
		b.module.AddFunction(fn)
	}
	b.currentFunc = fn
	return fn
}

// DeclareFunction declares an external function
func (b *Builder) DeclareFunction(name string, retType types.Type, params []types.Type, variadic bool) *ir.Function {
	fnType := types.NewFunction(retType, params, variadic)
	fn := ir.NewFunction(name, fnType)
	fn.Linkage = ir.ExternalLinkage
	if b.module != nil {
		b.module.AddFunction(fn)
	}
	return fn
}

// CreateGlobalVariable creates a global variable
func (b *Builder) CreateGlobalVariable(name string, typ types.Type, initializer ir.Constant) *ir.Global {
	g := &ir.Global{
		Initializer: initializer,
		Linkage:     ir.ExternalLinkage,
	}
	// Globals are pointers to the value
	g.SetType(types.NewPointer(typ))
	g.SetName(name)
	if b.module != nil {
		b.module.AddGlobal(g)
	}
	return g
}

// CreateGlobalConstant creates a global constant
func (b *Builder) CreateGlobalConstant(name string, initializer ir.Constant) *ir.Global {
	g := &ir.Global{
		Initializer: initializer,
		IsConstant:  true,
		Linkage:     ir.ExternalLinkage,
	}
	g.SetName(name)
	// Globals are pointers to the value
	g.SetType(types.NewPointer(initializer.Type()))
	if b.module != nil {
		b.module.AddGlobal(g)
	}
	return g
}

// ============================================================================
// Basic Block operations
// ============================================================================

// CreateBlock creates a new basic block
func (b *Builder) CreateBlock(name string) *ir.BasicBlock {
	block := ir.NewBasicBlock(name)
	if b.currentFunc != nil {
		b.currentFunc.AddBlock(block)
	}
	return block
}

// CreateBlockInFunction creates a block in a specific function
func (b *Builder) CreateBlockInFunction(name string, fn *ir.Function) *ir.BasicBlock {
	block := ir.NewBasicBlock(name)
	fn.AddBlock(block)
	return block
}

// ============================================================================
// Terminator instructions
// ============================================================================

// CreateRet creates a return instruction
func (b *Builder) CreateRet(v ir.Value) *ir.RetInst {
	inst := &ir.RetInst{}
	inst.Op = ir.OpRet
	if v != nil {
		inst.SetOperand(0, v)
	}
	b.insert(inst)
	return inst
}

// CreateRetVoid creates a void return
func (b *Builder) CreateRetVoid() *ir.RetInst {
	inst := &ir.RetInst{}
	inst.Op = ir.OpRet
	b.insert(inst)
	return inst
}

// CreateBr creates an unconditional branch
func (b *Builder) CreateBr(target *ir.BasicBlock) *ir.BrInst {
	inst := &ir.BrInst{Target: target}
	inst.Op = ir.OpBr
	b.insert(inst)
	// Update CFG
	b.currentBlock.Successors = append(b.currentBlock.Successors, target)
	target.Predecessors = append(target.Predecessors, b.currentBlock)
	return inst
}

// CreateCondBr creates a conditional branch
func (b *Builder) CreateCondBr(cond ir.Value, trueBlock, falseBlock *ir.BasicBlock) *ir.CondBrInst {
	inst := &ir.CondBrInst{
		Condition:  cond,
		TrueBlock:  trueBlock,
		FalseBlock: falseBlock,
	}
	inst.Op = ir.OpCondBr
	b.insert(inst)
	// Update CFG
	b.currentBlock.Successors = append(b.currentBlock.Successors, trueBlock, falseBlock)
	trueBlock.Predecessors = append(trueBlock.Predecessors, b.currentBlock)
	falseBlock.Predecessors = append(falseBlock.Predecessors, b.currentBlock)
	return inst
}

// CreateSwitch creates a switch instruction
func (b *Builder) CreateSwitch(cond ir.Value, defaultBlock *ir.BasicBlock, numCases int) *ir.SwitchInst {
	inst := &ir.SwitchInst{
		Condition:    cond,
		DefaultBlock: defaultBlock,
		Cases:        make([]ir.SwitchCase, 0, numCases),
	}
	inst.Op = ir.OpSwitch
	b.insert(inst)
	b.currentBlock.Successors = append(b.currentBlock.Successors, defaultBlock)
	defaultBlock.Predecessors = append(defaultBlock.Predecessors, b.currentBlock)
	return inst
}

// AddCase adds a case to a switch instruction
func (b *Builder) AddCase(sw *ir.SwitchInst, val *ir.ConstantInt, block *ir.BasicBlock) {
	sw.Cases = append(sw.Cases, ir.SwitchCase{Value: val, Block: block})
	// Update CFG
	parent := sw.Parent()
	if parent != nil {
		parent.Successors = append(parent.Successors, block)
		block.Predecessors = append(block.Predecessors, parent)
	}
}

// CreateUnreachable creates an unreachable instruction
func (b *Builder) CreateUnreachable() *ir.UnreachableInst {
	inst := &ir.UnreachableInst{}
	inst.Op = ir.OpUnreachable
	b.insert(inst)
	return inst
}

// ============================================================================
// Binary operations
// ============================================================================

func (b *Builder) createBinaryOp(op ir.Opcode, lhs, rhs ir.Value, name string) *ir.BinaryInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.BinaryInst{}
	inst.Op = op
	inst.SetName(name)
	inst.SetOperand(0, lhs)
	inst.SetOperand(1, rhs)
	// Implicitly set type to LHS type (standard for binary ops)
	inst.SetType(lhs.Type())
	b.insert(inst)
	return inst
}

// CreateAdd creates an add instruction
func (b *Builder) CreateAdd(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpAdd, lhs, rhs, name)
}

// CreateNSWAdd creates an add with no signed wrap
func (b *Builder) CreateNSWAdd(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpAdd, lhs, rhs, name)
	inst.NoSignedWrap = true
	return inst
}

// CreateNUWAdd creates an add with no unsigned wrap
func (b *Builder) CreateNUWAdd(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpAdd, lhs, rhs, name)
	inst.NoUnsignedWrap = true
	return inst
}

// CreateSub creates a sub instruction
func (b *Builder) CreateSub(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpSub, lhs, rhs, name)
}

// CreateNSWSub creates a sub with no signed wrap
func (b *Builder) CreateNSWSub(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpSub, lhs, rhs, name)
	inst.NoSignedWrap = true
	return inst
}

// CreateMul creates a mul instruction
func (b *Builder) CreateMul(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpMul, lhs, rhs, name)
}

// CreateNSWMul creates a mul with no signed wrap
func (b *Builder) CreateNSWMul(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpMul, lhs, rhs, name)
	inst.NoSignedWrap = true
	return inst
}

// CreateUDiv creates an unsigned division
func (b *Builder) CreateUDiv(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpUDiv, lhs, rhs, name)
}

// CreateExactUDiv creates an exact unsigned division
func (b *Builder) CreateExactUDiv(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpUDiv, lhs, rhs, name)
	inst.Exact = true
	return inst
}

// CreateSDiv creates a signed division
func (b *Builder) CreateSDiv(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpSDiv, lhs, rhs, name)
}

// CreateExactSDiv creates an exact signed division
func (b *Builder) CreateExactSDiv(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	inst := b.createBinaryOp(ir.OpSDiv, lhs, rhs, name)
	inst.Exact = true
	return inst
}

// CreateURem creates unsigned remainder
func (b *Builder) CreateURem(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpURem, lhs, rhs, name)
}

// CreateSRem creates signed remainder
func (b *Builder) CreateSRem(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpSRem, lhs, rhs, name)
}

// Floating point operations
func (b *Builder) CreateFAdd(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpFAdd, lhs, rhs, name)
}

func (b *Builder) CreateFSub(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpFSub, lhs, rhs, name)
}

func (b *Builder) CreateFMul(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpFMul, lhs, rhs, name)
}

func (b *Builder) CreateFDiv(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpFDiv, lhs, rhs, name)
}

func (b *Builder) CreateFRem(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpFRem, lhs, rhs, name)
}

// ============================================================================
// Bitwise operations
// ============================================================================

func (b *Builder) CreateShl(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpShl, lhs, rhs, name)
}

func (b *Builder) CreateLShr(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpLShr, lhs, rhs, name)
}

func (b *Builder) CreateAShr(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpAShr, lhs, rhs, name)
}

func (b *Builder) CreateAnd(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpAnd, lhs, rhs, name)
}

func (b *Builder) CreateOr(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpOr, lhs, rhs, name)
}

func (b *Builder) CreateXor(lhs, rhs ir.Value, name string) *ir.BinaryInst {
	return b.createBinaryOp(ir.OpXor, lhs, rhs, name)
}

// ============================================================================
// Memory operations
// ============================================================================

// CreateAlloca creates a stack allocation
func (b *Builder) CreateAlloca(typ types.Type, name string) *ir.AllocaInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.AllocaInst{
		AllocatedType: typ,
	}
	inst.Op = ir.OpAlloca
	inst.SetType(types.NewPointer(typ))
	inst.SetName(name)
	b.insert(inst)
	return inst
}

// CreateAllocaWithCount creates an array allocation on stack
func (b *Builder) CreateAllocaWithCount(typ types.Type, count ir.Value, name string) *ir.AllocaInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.AllocaInst{
		AllocatedType: typ,
		NumElements:   count,
	}
	inst.Op = ir.OpAlloca
	inst.SetType(types.NewPointer(typ))
	inst.SetName(name)
	b.insert(inst)
	return inst
}

// CreateLoad creates a load instruction
func (b *Builder) CreateLoad(typ types.Type, ptr ir.Value, name string) *ir.LoadInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.LoadInst{}
	inst.Op = ir.OpLoad
	inst.SetName(name)
	inst.SetType(typ)
	inst.SetOperand(0, ptr)
	b.insert(inst)
	return inst
}

// CreateVolatileLoad creates a volatile load
func (b *Builder) CreateVolatileLoad(typ types.Type, ptr ir.Value, name string) *ir.LoadInst {
	inst := b.CreateLoad(typ, ptr, name)
	inst.Volatile = true
	return inst
}

// CreateAlignedLoad creates an aligned load
func (b *Builder) CreateAlignedLoad(typ types.Type, ptr ir.Value, align int, name string) *ir.LoadInst {
	inst := b.CreateLoad(typ, ptr, name)
	inst.Alignment = align
	return inst
}

// CreateStore creates a store instruction
func (b *Builder) CreateStore(val ir.Value, ptr ir.Value) *ir.StoreInst {
	inst := &ir.StoreInst{}
	inst.Op = ir.OpStore
	inst.SetOperand(0, val)
	inst.SetOperand(1, ptr)
	b.insert(inst)
	return inst
}

// CreateVolatileStore creates a volatile store
func (b *Builder) CreateVolatileStore(val ir.Value, ptr ir.Value) *ir.StoreInst {
	inst := b.CreateStore(val, ptr)
	inst.Volatile = true
	return inst
}

// CreateAlignedStore creates an aligned store
func (b *Builder) CreateAlignedStore(val ir.Value, ptr ir.Value, align int) *ir.StoreInst {
	inst := b.CreateStore(val, ptr)
	inst.Alignment = align
	return inst
}

// CreateGEP creates a getelementptr instruction
func (b *Builder) CreateGEP(pointeeType types.Type, ptr ir.Value, indices []ir.Value, name string) *ir.GetElementPtrInst {
	if name == "" {
		name = b.generateName()
	}
	operands := make([]ir.Value, 1+len(indices))
	operands[0] = ptr
	copy(operands[1:], indices)

	inst := &ir.GetElementPtrInst{
		SourceElementType: pointeeType,
	}
	inst.Op = ir.OpGetElementPtr
	inst.SetName(name)
	// Simplified return type calculation
	inst.SetType(types.NewPointer(pointeeType))
	for i, op := range operands {
		inst.SetOperand(i, op)
	}
	b.insert(inst)
	return inst
}

// CreateInBoundsGEP creates an inbounds GEP
func (b *Builder) CreateInBoundsGEP(pointeeType types.Type, ptr ir.Value, indices []ir.Value, name string) *ir.GetElementPtrInst {
	inst := b.CreateGEP(pointeeType, ptr, indices, name)
	inst.InBounds = true
	return inst
}

// CreateStructGEP creates a GEP to access a struct field
func (b *Builder) CreateStructGEP(structType types.Type, ptr ir.Value, idx int, name string) *ir.GetElementPtrInst {
	zero := b.ConstInt(types.I32, 0)
	idxVal := b.ConstInt(types.I32, int64(idx))
	return b.CreateGEP(structType, ptr, []ir.Value{zero, idxVal}, name)
}

// ============================================================================
// Cast operations
// ============================================================================

func (b *Builder) createCast(op ir.Opcode, v ir.Value, destTy types.Type, name string) *ir.CastInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.CastInst{
		DestType: destTy,
	}
	inst.Op = op
	inst.SetName(name)
	inst.SetType(destTy)
	inst.SetOperand(0, v)
	b.insert(inst)
	return inst
}

func (b *Builder) CreateTrunc(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpTrunc, v, destTy, name)
}

func (b *Builder) CreateZExt(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpZExt, v, destTy, name)
}

func (b *Builder) CreateSExt(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpSExt, v, destTy, name)
}

func (b *Builder) CreateFPTrunc(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpFPTrunc, v, destTy, name)
}

func (b *Builder) CreateFPExt(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpFPExt, v, destTy, name)
}

func (b *Builder) CreateFPToUI(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpFPToUI, v, destTy, name)
}

func (b *Builder) CreateFPToSI(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpFPToSI, v, destTy, name)
}

func (b *Builder) CreateUIToFP(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpUIToFP, v, destTy, name)
}

func (b *Builder) CreateSIToFP(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpSIToFP, v, destTy, name)
}

func (b *Builder) CreatePtrToInt(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpPtrToInt, v, destTy, name)
}

func (b *Builder) CreateIntToPtr(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpIntToPtr, v, destTy, name)
}

func (b *Builder) CreateBitCast(v ir.Value, destTy types.Type, name string) *ir.CastInst {
	return b.createCast(ir.OpBitcast, v, destTy, name)
}

// ============================================================================
// Comparison operations
// ============================================================================

func (b *Builder) CreateICmp(pred ir.ICmpPredicate, lhs, rhs ir.Value, name string) *ir.ICmpInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.ICmpInst{
		Predicate: pred,
	}
	inst.Op = ir.OpICmp
	inst.SetName(name)
	inst.SetType(types.I1)
	inst.SetOperand(0, lhs)
	inst.SetOperand(1, rhs)
	b.insert(inst)
	return inst
}

func (b *Builder) CreateFCmp(pred ir.FCmpPredicate, lhs, rhs ir.Value, name string) *ir.FCmpInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.FCmpInst{
		Predicate: pred,
	}
	inst.Op = ir.OpFCmp
	inst.SetName(name)
	inst.SetType(types.I1)
	inst.SetOperand(0, lhs)
	inst.SetOperand(1, rhs)
	b.insert(inst)
	return inst
}

// Convenience comparison methods
func (b *Builder) CreateICmpEQ(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpEQ, lhs, rhs, name)
}

func (b *Builder) CreateICmpNE(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpNE, lhs, rhs, name)
}

func (b *Builder) CreateICmpSLT(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpSLT, lhs, rhs, name)
}

func (b *Builder) CreateICmpSLE(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpSLE, lhs, rhs, name)
}

func (b *Builder) CreateICmpSGT(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpSGT, lhs, rhs, name)
}

func (b *Builder) CreateICmpSGE(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpSGE, lhs, rhs, name)
}

func (b *Builder) CreateICmpULT(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpULT, lhs, rhs, name)
}

func (b *Builder) CreateICmpULE(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpULE, lhs, rhs, name)
}

func (b *Builder) CreateICmpUGT(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpUGT, lhs, rhs, name)
}

func (b *Builder) CreateICmpUGE(lhs, rhs ir.Value, name string) *ir.ICmpInst {
	return b.CreateICmp(ir.ICmpUGE, lhs, rhs, name)
}

// ============================================================================
// Other operations
// ============================================================================

// CreatePhi creates a phi node
func (b *Builder) CreatePhi(typ types.Type, name string) *ir.PhiInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.PhiInst{}
	inst.Op = ir.OpPhi
	inst.SetName(name)
	inst.SetType(typ)
	b.insert(inst)
	return inst
}

// CreateSelect creates a select instruction
func (b *Builder) CreateSelect(cond ir.Value, trueVal, falseVal ir.Value, name string) *ir.SelectInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.SelectInst{}
	inst.Op = ir.OpSelect
	inst.SetName(name)
	inst.SetType(trueVal.Type())
	inst.SetOperand(0, cond)
	inst.SetOperand(1, trueVal)
	inst.SetOperand(2, falseVal)
	b.insert(inst)
	return inst
}

// CreateCall creates a function call
func (b *Builder) CreateCall(fn *ir.Function, args []ir.Value, name string) *ir.CallInst {
	if name == "" && fn.FuncType.ReturnType.Kind() != types.VoidKind {
		name = b.generateName()
	}
	inst := &ir.CallInst{
		Callee: fn,
	}
	inst.Op = ir.OpCall
	inst.SetName(name)
	inst.SetType(fn.FuncType.ReturnType)
	for i, arg := range args {
		inst.SetOperand(i, arg)
	}
	b.insert(inst)
	return inst
}

// CreateCallByName creates a call to a named function
func (b *Builder) CreateCallByName(name string, retType types.Type, args []ir.Value, resultName string) *ir.CallInst {
	if resultName == "" && retType.Kind() != types.VoidKind {
		resultName = b.generateName()
	}
	inst := &ir.CallInst{
		CalleeName: name,
	}
	inst.Op = ir.OpCall
	inst.SetName(resultName)
	inst.SetType(retType)
	for i, arg := range args {
		inst.SetOperand(i, arg)
	}
	b.insert(inst)
	return inst
}

// CreateExtractValue extracts a value from an aggregate
func (b *Builder) CreateExtractValue(agg ir.Value, indices []int, name string) *ir.ExtractValueInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.ExtractValueInst{
		Indices: indices,
	}
	inst.Op = ir.OpExtractValue
	inst.SetName(name)
	inst.SetType(agg.Type()) // Approximation
	inst.SetOperand(0, agg)
	b.insert(inst)
	return inst
}

// CreateInsertValue inserts a value into an aggregate
func (b *Builder) CreateInsertValue(agg ir.Value, val ir.Value, indices []int, name string) *ir.InsertValueInst {
	if name == "" {
		name = b.generateName()
	}
	inst := &ir.InsertValueInst{
		Indices: indices,
	}
	inst.Op = ir.OpInsertValue
	inst.SetName(name)
	inst.SetType(agg.Type())
	inst.SetOperand(0, agg)
	inst.SetOperand(1, val)
	b.insert(inst)
	return inst
}

// ============================================================================
// Constant creation
// ============================================================================

// ConstInt creates an integer constant
func (b *Builder) ConstInt(typ *types.IntType, val int64) *ir.ConstantInt {
	c := &ir.ConstantInt{
		Value: val,
	}
	c.SetType(typ)
	return c
}

// ConstFloat creates a float constant
func (b *Builder) ConstFloat(typ *types.FloatType, val float64) *ir.ConstantFloat {
	c := &ir.ConstantFloat{
		Value: val,
	}
	c.SetType(typ)
	return c
}

// ConstNull creates a null pointer constant
func (b *Builder) ConstNull(ptrType *types.PointerType) *ir.ConstantNull {
	c := &ir.ConstantNull{}
	c.SetType(ptrType)
	return c
}

// ConstUndef creates an undefined value
func (b *Builder) ConstUndef(typ types.Type) *ir.ConstantUndef {
	c := &ir.ConstantUndef{}
	c.SetType(typ)
	return c
}

// ConstZero creates a zero initializer
func (b *Builder) ConstZero(typ types.Type) *ir.ConstantZero {
	c := &ir.ConstantZero{}
	c.SetType(typ)
	return c
}

// True returns i1 1
func (b *Builder) True() *ir.ConstantInt {
	return b.ConstInt(types.I1, 1)
}

// False returns i1 0
func (b *Builder) False() *ir.ConstantInt {
	return b.ConstInt(types.I1, 0)
}