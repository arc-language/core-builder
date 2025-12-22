package main

import (
	"fmt"

	"github.com/arc-language/core-builder/builder"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
)

func main() {
	runSample("Factorial (Recursion)", buildFactorialModule)
	runSample("Fibonacci (Recursion)", buildFibonacciModule)
	runSample("GCD (Loop with Phi)", buildGCDModule)
	runSample("Structs & Pointers", buildStructModule)
	runSample("Switch Statement", buildSwitchModule)
	runSample("Arrays & Globals", buildGlobalArrayModule)
}

func runSample(name string, builderFunc func() *ir.Module) {
	fmt.Println("================================================================================")
	fmt.Printf("SAMPLE: %s\n", name)
	fmt.Println("================================================================================")
	mod := builderFunc()
	fmt.Println(mod.String())
	fmt.Println()
}

// ============================================================================
// Sample 1: Factorial (Basic Recursion)
// ============================================================================
func buildFactorialModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("factorial_sample")

	// func factorial(n i32) i32
	fn := b.CreateFunction("factorial", types.I32, []types.Type{types.I32}, false)
	n := fn.Arguments[0]
	n.SetName("n")

	entry := b.CreateBlock("entry")
	thenBB := b.CreateBlock("then")
	elseBB := b.CreateBlock("else")

	b.SetInsertPoint(entry)
	one := b.ConstInt(types.I32, 1)
	cond := b.CreateICmpSLE(n, one, "cmp")
	b.CreateCondBr(cond, thenBB, elseBB)

	b.SetInsertPoint(thenBB)
	b.CreateRet(one)

	b.SetInsertPoint(elseBB)
	nMinus1 := b.CreateSub(n, one, "sub")
	call := b.CreateCall(fn, []ir.Value{nMinus1}, "call")
	result := b.CreateMul(n, call, "mul")
	b.CreateRet(result)

	return mod
}

// ============================================================================
// Sample 2: Fibonacci (Recursive with multiple branches)
// ============================================================================
func buildFibonacciModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("fibonacci_sample")

	// func fib(n i32) i32
	fn := b.CreateFunction("fib", types.I32, []types.Type{types.I32}, false)
	n := fn.Arguments[0]
	n.SetName("n")

	entry := b.CreateBlock("entry")
	baseCase := b.CreateBlock("base_case")
	recurse := b.CreateBlock("recurse")

	b.SetInsertPoint(entry)
	two := b.ConstInt(types.I32, 2)
	// if n < 2
	cond := b.CreateICmpSLT(n, two, "cond")
	b.CreateCondBr(cond, baseCase, recurse)

	b.SetInsertPoint(baseCase)
	b.CreateRet(n)

	b.SetInsertPoint(recurse)
	one := b.ConstInt(types.I32, 1)
	
	// fib(n-1)
	sub1 := b.CreateSub(n, one, "sub1")
	call1 := b.CreateCall(fn, []ir.Value{sub1}, "call1")
	
	// fib(n-2)
	sub2 := b.CreateSub(n, two, "sub2")
	call2 := b.CreateCall(fn, []ir.Value{sub2}, "call2")
	
	// result = call1 + call2
	sum := b.CreateAdd(call1, call2, "sum")
	b.CreateRet(sum)

	return mod
}

// ============================================================================
// Sample 3: GCD (Looping with Phi Nodes)
// ============================================================================
func buildGCDModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("gcd_sample")

	// func gcd(a, b i32) i32
	fn := b.CreateFunction("gcd", types.I32, []types.Type{types.I32, types.I32}, false)
	aArg := fn.Arguments[0]
	aArg.SetName("a")
	bArg := fn.Arguments[1]
	bArg.SetName("b")

	entry := b.CreateBlock("entry")
	loopHead := b.CreateBlock("loop.head")
	loopBody := b.CreateBlock("loop.body")
	exit := b.CreateBlock("exit")

	// Entry: Jump directly to loop head
	b.SetInsertPoint(entry)
	b.CreateBr(loopHead)

	// Loop Head: Phi nodes for 'a' and 'b'
	b.SetInsertPoint(loopHead)
	
	phiA := b.CreatePhi(types.I32, "curr_a")
	phiB := b.CreatePhi(types.I32, "curr_b")
	
	// Populate initial incoming values (from entry)
	phiA.AddIncoming(aArg, entry)
	phiB.AddIncoming(bArg, entry)

	// Check if b == 0
	zero := b.ConstInt(types.I32, 0)
	cond := b.CreateICmpNE(phiB, zero, "cond")
	b.CreateCondBr(cond, loopBody, exit)

	// Loop Body: t = b; b = a % b; a = t;
	b.SetInsertPoint(loopBody)
	rem := b.CreateSRem(phiA, phiB, "rem")
	
	// Loop back to head with new values:
	// new 'a' is old 'b' (phiB)
	// new 'b' is 'rem'
	phiA.AddIncoming(phiB, loopBody)
	phiB.AddIncoming(rem, loopBody)
	b.CreateBr(loopHead)

	// Exit: return 'a' (which is in phiA)
	b.SetInsertPoint(exit)
	b.CreateRet(phiA)

	return mod
}

// ============================================================================
// Sample 4: Structs (GEP, Load, Store)
// ============================================================================
func buildStructModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("struct_sample")

	// type Point struct { x, y i32 }
	structTy := types.NewStruct("Point", []types.Type{types.I32, types.I32}, false)
	mod.Types["Point"] = structTy

	// func update_y(p *Point, new_y i32)
	ptrTy := types.NewPointer(structTy)
	fn := b.CreateFunction("update_y", types.Void, []types.Type{ptrTy, types.I32}, false)
	p := fn.Arguments[0]
	p.SetName("p")
	newY := fn.Arguments[1]
	newY.SetName("new_y")

	entry := b.CreateBlock("entry")
	b.SetInsertPoint(entry)

	// Get pointer to field 1 (y)
	// GEP indices: 0 (dereference pointer), 1 (field index)
	zero := b.ConstInt(types.I32, 0)
	one := b.ConstInt(types.I32, 1)
	
	// gep %p, 0, 1
	gep := b.CreateGEP(structTy, p, []ir.Value{zero, one}, "y_ptr")
	
	// store %new_y, %y_ptr
	b.CreateStore(newY, gep)
	
	b.CreateRetVoid()

	return mod
}

// ============================================================================
// Sample 5: Switch Statement
// ============================================================================
func buildSwitchModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("switch_sample")

	// func classify(n i32) i32
	fn := b.CreateFunction("classify", types.I32, []types.Type{types.I32}, false)
	n := fn.Arguments[0]
	n.SetName("n")

	entry := b.CreateBlock("entry")
	caseZero := b.CreateBlock("case_zero")
	caseOne := b.CreateBlock("case_one")
	defaultBB := b.CreateBlock("default")
	merge := b.CreateBlock("merge")

	b.SetInsertPoint(entry)
	sw := b.CreateSwitch(n, defaultBB, 2)
	b.AddCase(sw, b.ConstInt(types.I32, 0), caseZero)
	b.AddCase(sw, b.ConstInt(types.I32, 1), caseOne)

	// Case 0
	b.SetInsertPoint(caseZero)
	res0 := b.ConstInt(types.I32, 100)
	b.CreateBr(merge)

	// Case 1
	b.SetInsertPoint(caseOne)
	res1 := b.ConstInt(types.I32, 200)
	b.CreateBr(merge)

	// Default
	b.SetInsertPoint(defaultBB)
	resDef := b.ConstInt(types.I32, -1)
	b.CreateBr(merge)

	// Merge
	b.SetInsertPoint(merge)
	phi := b.CreatePhi(types.I32, "result")
	phi.AddIncoming(res0, caseZero)
	phi.AddIncoming(res1, caseOne)
	phi.AddIncoming(resDef, defaultBB)
	
	b.CreateRet(phi)

	return mod
}

// ============================================================================
// Sample 6: Globals and Arrays
// ============================================================================
func buildGlobalArrayModule() *ir.Module {
	b := builder.New()
	mod := b.CreateModule("global_array_sample")

	// global g_val = 42
	gVal := b.CreateGlobalVariable("g_val", types.I32, b.ConstInt(types.I32, 42))

	// func main() i32
	_ = b.CreateFunction("main", types.I32, nil, false) // FIX: Discard unused variable
	
	entry := b.CreateBlock("entry")
	b.SetInsertPoint(entry)

	// Load global
	val := b.CreateLoad(types.I32, gVal, "loaded_val")

	// Stack array: [2 x i32]
	arrTy := types.NewArray(types.I32, 2)
	arr := b.CreateAlloca(arrTy, "stack_arr")

	// Store val into index 0
	zero := b.ConstInt(types.I32, 0)
	// GEP: &arr[0][0]
	elemPtr := b.CreateGEP(arrTy, arr, []ir.Value{zero, zero}, "elem_ptr")
	b.CreateStore(val, elemPtr)

	b.CreateRet(val)

	return mod
}