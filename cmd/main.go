// Example usage of github.com/arc-language/core-builder
//
// This demonstrates building IR for a simple factorial function:
//
//   int factorial(int n) {
//       if (n <= 1) return 1;
//       return n * factorial(n - 1);
//   }
//
package main

import (
	"fmt"
	
	builder "github.com/arc-language/core-builder"
	"github.com/arc-language/core-builder/ir"
	"github.com/arc-language/core-builder/types"
)

func main() {
	// Create a new builder and module
	b := builder.New()
	mod := b.CreateModule("factorial_example")

	// Build factorial function
	buildFactorial(b)

	// Build a main function that calls factorial
	buildMain(b)

	// Print the generated IR
	fmt.Println(mod.String())
}

func buildFactorial(b *builder.Builder) *ir.Function {
	// Create function: i32 @factorial(i32 %n)
	fn := b.CreateFunction("factorial", types.I32, []types.Type{types.I32}, false)
	fn.Arguments[0].SetName("n")

	// Create basic blocks
	entry := b.CreateBlock("entry")
	thenBB := b.CreateBlock("then")
	elseBB := b.CreateBlock("else")

	// Entry block: compare n <= 1
	b.SetInsertPoint(entry)
	n := fn.Arguments[0]
	one := b.ConstInt(types.I32, 1)
	cond := b.CreateICmpSLE(n, one, "cmp")
	b.CreateCondBr(cond, thenBB, elseBB)

	// Then block: return 1
	b.SetInsertPoint(thenBB)
	b.CreateRet(one)

	// Else block: return n * factorial(n - 1)
	b.SetInsertPoint(elseBB)
	nMinus1 := b.CreateSub(n, one, "n_minus_1")
	recursiveCall := b.CreateCall(fn, []ir.Value{nMinus1}, "rec_result")
	result := b.CreateMul(n, recursiveCall, "result")
	b.CreateRet(result)

	return fn
}

func buildMain(b *builder.Builder) *ir.Function {
	// Create function: i32 @main()
	fn := b.CreateFunction("main", types.I32, nil, false)

	entry := b.CreateBlock("entry")
	b.SetInsertPoint(entry)

	// Get factorial function
	factorial := b.Module().GetFunction("factorial")

	// Call factorial(5)
	five := b.ConstInt(types.I32, 5)
	result := b.CreateCall(factorial, []ir.Value{five}, "fact_result")

	// Return the result
	b.CreateRet(result)

	return fn
}