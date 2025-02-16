package main

import (
	"fmt"
)

type LExpr interface {
	LPrint() string
	LAbstract(l LVar) LExpr
	Copy() LExpr
	LApply(b LVar, replace LExpr) LExpr
}

// If can come in bracketed and unbracketed variants, must change LAbstract method
// If only bracketed, 1 single consistent LAbstract method
// Right now, assuming that LambdaExpression comes in only as Bracketed Variant

// Handles Abstracted expr via: Binding = some LVar
type LambdaExpression struct {
	Binding LVar
	Exprs   []LExpr
	Repr    string
}

//func (l *LambdaExpression) LPrint() string {
//	print_l := ""
//	if l.Binding.LPrint() != "" {
//		print_l += "λ" + l.Binding.LPrint() + "."
//	}
//	print_l += "("
//	for _, expr := range l.Exprs {
//		print_l += expr.LPrint()
//	}
//	print_l += ")"
//	return print_l
//}

func (l *LambdaExpression) LPrint() string {
	return l.Repr
}

func (l *LambdaExpression) LAbstract(b LVar) LExpr {
	out_exprs := []LExpr{l}
	out_str := "λ" + b.LPrint() + "." + l.LPrint()

	out_le := LambdaExpression{
		Binding: b,
		Exprs:   out_exprs,
		Repr:    out_str,
	}

	return &out_le
}

func (l *LambdaExpression) Copy() LExpr {
	copy_expr := make([]LExpr, len(l.Exprs))
	for i, expr := range l.Exprs {
		copy_expr[i] = expr.Copy()
	}
	return &LambdaExpression{
		Binding: LVar{Symbol: l.Binding.Symbol},
		Exprs:   copy_expr,
		Repr:    l.Repr,
	}
}

func (l *LambdaExpression) LApply(b LVar, replace LExpr) LExpr {
	new_exprs := make([]LExpr, len(l.Exprs))
	new_repr := "λ" + l.Binding.LPrint() + ".("
	for i, expr := range l.Exprs {
		new_exprs[i] = expr.LApply(b, replace)
		new_repr += new_exprs[i].LPrint()
	}
	new_repr += ")"
	return &LambdaExpression{
		Binding: LVar{Symbol: l.Binding.Symbol},
		Exprs:   new_exprs,
		Repr:    new_repr,
	}

}

func LApplyInit(l1 LambdaExpression, l2 LExpr) LambdaExpression {
	new_exprs := make([]LExpr, len(l1.Exprs))
	new_repr := "("
	for i, expr := range l1.Exprs {
		new_exprs[i] = expr.LApply(l1.Binding, l2)
		new_repr += new_exprs[i].LPrint()
	}
	new_repr += ")"
	return LambdaExpression{
		Binding: LVar{Symbol: ""},
		Exprs:   new_exprs,
		Repr:    new_repr,
	}

}

func ConcatenateLExprs(lexprs []LExpr) LambdaExpression {
	new_repr := "("
	for _, expr := range lexprs {
		new_repr += expr.LPrint()
	}
	new_repr += ")"
	return LambdaExpression{
		Binding: LVar{},
		Exprs:   lexprs,
		Repr:    new_repr,
	}
}

type LVar struct {
	Symbol string
}

func (l *LVar) LPrint() string {
	return l.Symbol
}

func (l *LVar) LAbstract(b LVar) LExpr {
	Exprs := []LExpr{}
	Exprs = append(Exprs, l)
	out_str := "λ" + b.LPrint() + ".(" + l.LPrint() + ")"
	out_expr := LambdaExpression{
		Binding: b,
		Exprs:   Exprs,
		Repr:    out_str,
	}

	return &out_expr
}

func (l *LVar) Copy() LExpr {
	return &LVar{Symbol: l.Symbol}
}

func (l *LVar) LApply(b LVar, replace LExpr) LExpr {
	if l.LPrint() == b.LPrint() {
		return b.Copy()
	} else {
		return l.Copy()
	}
}

func main() {
	fmt.Println("yo")

	x := LVar{Symbol: "x"}
	fmt.Println(x.LPrint())
	fmt.Println(x)

	byx := x.LAbstract(LVar{Symbol: "y"})
	fmt.Println("Form of byx: ", byx.LPrint())
	fmt.Println(byx)

	bzbyx := ConcatenateLExprs([]LExpr{byx, &x})
	fmt.Println("Print bzbyx: ", bzbyx.LPrint())
	fmt.Println(bzbyx)
}
