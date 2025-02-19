package main

import (
	"errors"
	"fmt"
	"strconv"
)

// NOTE: For copying: λ

/*
	 LExpr - Lambda Expression type
		- allowing variables, abstractions, and compounding of these terms

LExpr requires the following:
  - LPrint() - Expresses the LExpr as a string in it's most simple form
    (an expression of the form "((x))" should return the string "x")
    such that LPrint's value is shared by equivalent terms like (x), x, ((x)), etc
  - LAbstract(l LVar) - for a lmabda expression e, given some lambda variable a,
    returns the result λa.e. any instances of variable a in e are now bound to
    a, i.e. if we give λa.e an input b, (λa.e)(b), then we would return e' where
    all the a values in e have been replaced by b
    λa.e(b) = e' = e[a => b]
  - Copy() - Generate a new copy of LExpr
  - LApply(b LVar, replace LEXpr) LExpr - For an LExpr e, executes e(replace) if e is
    some LExpression of the form e = λa.e2 and a=b (b arg added for recursion purposes)
    If e is some LVar, directly replaces e with replace value if e=b
  - LEquals(l2 LExpr) bool - for l1.LEquals(l2), checks if l1 and l2 are in the same
    equivalence class (if l1.LPrint() == l2.LPrint())
*/
type LExpr interface {
	LPrint() string
	LAbstract(l LVar) LExpr
	Copy() LExpr
	LApply(b LVar, replace LExpr) LExpr
	LEquals(l2 LExpr) bool
}

// If can come in bracketed and unbracketed variants, must change LAbstract method
// If only bracketed, 1 single consistent LAbstract method
// Right now, assuming that LExpression comes in only as Bracketed Variant

// Handles Abstracted expr via: Binding = some LVar
type LExpression struct {
	Binding LVar
	Exprs   []LExpr
	Repr    string
}

func (l *LExpression) LPrint() string {
	if (len(l.Binding.Symbol) == 0) && (len(l.Exprs) == 1) {
		return l.Exprs[0].LPrint()
	} else {
		return l.Repr
	}
}

func (l *LExpression) LAbstract(b LVar) LExpr {
	out_exprs := []LExpr{l}
	out_str := "λ" + b.LPrint() + "." + l.LPrint()

	out_le := LExpression{
		Binding: b,
		Exprs:   out_exprs,
		Repr:    out_str,
	}

	return &out_le
}

func (l *LExpression) Copy() LExpr {
	copy_expr := make([]LExpr, len(l.Exprs))
	for i, expr := range l.Exprs {
		copy_expr[i] = expr.Copy()
	}
	return &LExpression{
		Binding: LVar{Symbol: l.Binding.Symbol},
		Exprs:   copy_expr,
		Repr:    l.Repr,
	}
}

func (l *LExpression) LApply(b LVar, replace LExpr) LExpr {
	new_exprs := make([]LExpr, len(l.Exprs))
	new_repr := "λ" + l.Binding.LPrint() + ".("
	for i, expr := range l.Exprs {
		new_exprs[i] = expr.LApply(b, replace)
		new_repr += new_exprs[i].LPrint()
	}
	new_repr += ")"
	return &LExpression{
		Binding: LVar{Symbol: l.Binding.Symbol},
		Exprs:   new_exprs,
		Repr:    new_repr,
	}

}

func (l *LExpression) LEquals(l2 LExpr) bool {
	return l.LPrint() == l2.LPrint()
}

func LApplyInit(l1 LExpression, l2 LExpr) LExpression {
	new_exprs := make([]LExpr, len(l1.Exprs))
	new_repr := "("
	for i, expr := range l1.Exprs {
		new_exprs[i] = expr.LApply(l1.Binding, l2)
		new_repr += new_exprs[i].LPrint()
	}
	new_repr += ")"
	return LExpression{
		Binding: LVar{Symbol: ""},
		Exprs:   new_exprs,
		Repr:    new_repr,
	}

}

func ConcatenateLExprs(lexprs []LExpr) LExpression {
	new_repr := "("
	for _, expr := range lexprs {
		new_repr += expr.LPrint()
	}
	new_repr += ")"
	return LExpression{
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
	out_expr := LExpression{
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

func (l *LVar) LEquals(l2 LExpr) bool {
	return l.LPrint() == l2.LPrint()
}

// Key symbols which should determine the control flow of parsing:
// (, ., ), and lambda
// ((()))
// ()()()
// L_.()(_L_.())
// (La.(bc)(xLy.(xz)))
// L      )
//    .(
//   a  bc

// base case: par(a) => LVar{a}
// LE: par(Lx.(y)) => LExpression{par(x), par(y)} => LamdaExpression{ LVar{x}, par(y)}
// For (a(_)La.(_)a)(bb)(cc)
// (a(_)La.(_)a)  (bb)  (cc)
// From (a(_)La.(_)a) and empty []LExpr we get
// if L prepends () => generate LE{LVar{x}, []}
// else, from () generate LE{LVar{""}, []}
// HOW DO YOU HAVE A UNIFIED METHOD OF HANDLING BOTH () AND NON-() AHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH
// MAYBE, creation of an LE should be equivalent to unwrapping -- Assume that when function is returning []LExpr, it's always returning for filling into a pre-existing []LExpr slice (input of function should at top level include an empty []LExpr slice)

// Recursively -- have funciton signature as (input string) []expr
// move pointer along input:
// if char not L or (: add LVar{} to []LExpr
// else, send (___) contents of parentheses into func recursively and then assign the []Lexpr as the contents of a LE
// Now trivially handles (aaa) and aaa by returning => LExpression{LVar{""}, [LVar{a}, LVar{a}, LVar{a}]},  [LVar{a}, LVar{a}, LVar{a}]
// INSTINCTS WERE CORRECT => This creates the need for equality testing being not direclty equality but equivalence classes ("(aaa)" and "aaa" are equal even though 1 has an unnecessary LExpression wrapper)
// How does recursive equality testing work in Go? Simple out is if LExpr.LPrint() == LExpr.LPrint() (with wrapping either in an extra pair of parentheses () since that is a syntactic but not semantic change) then equality
// Function would need to take both LExpr and []LExpr (for []LExpr, would need to wrap into LExpression arbitrarily)
// NOTE: Recursive function would have to, at end, have it's output wrapped into 1 final unnecessary LExpression for safety if is not already wrapped up as unnecessary LExpression (do you need to reduce down (x), ((x)), (((x))) for equality check? How would you do so?
// For wrapping up output, need to intelligently detect if already wrapped in final () and convert directly into LExpression or if need to create Repr via
// ( + x1.LPrint() + x2.LPrint() + ... + xn.LPrint() + )
// Guess could just check if slice returned by function is 1 length (and not a lambda expression) and if so doesn't need extra wrapping, else if >1 length or just 1 function lambda then needs extra wrapping

// Goal: Successfully parse (a(b)Lc.(c)(d(e))f)
type ParenTracker struct {
	Counter int
}

func (p *ParenTracker) Update(s string) {
	if s == "(" {
		p.Counter += 1
	} else if s == ")" {
		p.Counter -= 1
	}
}

func IsCapLetter(s string) bool {
	for _, r := range s {
		if (r < 'A') || (r > 'Z') {
			return false
		}
	}
	return true
}

// Checks if string is numeric digit
// nil err => conversion to int successful
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type ParseKeys struct {
	PState  string
	NextStr string
}

type LambdaParser struct {
	SrcStr             string
	PState             string
	ParenthesesTracker ParenTracker
	LambdaBinding      string
	CollectionStr      string
	LExprArr           []LExpr
}

func CreateParser(s string) LambdaParser {
	return LambdaParser{
		SrcStr:             s,
		PState:             "N",
		ParenthesesTracker: ParenTracker{Counter: 0},
		LambdaBinding:      "",
		CollectionStr:      "",
		LExprArr:           []LExpr{},
	}
}

func (lp *LambdaParser) GetParseKey(s string) ParseKeys {
	return ParseKeys{PState: lp.PState, NextStr: s}
}

func (lp *LambdaParser) PParse(s string) error {
	// While ParenthesesTracker.Counter != 0, collect str into CollectionStr
	// When hitting ), create new parser with CollectionStr and get resultant []LExpr
	// Bind []LExpr to LambdaBinding if present, create LExpression and append to LExprArr
	// NOTE: Clean out LambdaBinding, CollectionStr, check ParenthesesTracker == 0,
	// switch to N state
	var p_err error
	p_err = nil

	if (lp.PState == "N") && (s == "(") {
		lp.PState = "P"
	} else if (lp.PState == "P") && (s != ")") {
		lp.CollectionStr += s
	} else if (lp.PState == "P") && (s == ")") {
		nested_parser := CreateParser(lp.CollectionStr)
		parsed_contents, nest_err := nested_parser.DriveParse()
		if nest_err != nil {
			// TODO: Wrap nest_err in more information
			return nest_err
		}
		var new_lexpr LExpr
		new_lexpr = ConcatenateLExprs(parsed_contents)
		if len(lp.LambdaBinding) != 0 {
			new_lexpr = new_lexpr.LAbstract(LVar{lp.LambdaBinding})
		}
		lp.LExprArr = append(lp.LExprArr, new_lexpr)
	} else {
		state := lp.PState + " Parentheses-Enclosed Expr"
		p_err = StateCharError(state, s)
	}
	return p_err
}

func (lp *LambdaParser) LParse(s string) error {
	var l_err error
	l_err = nil
	if (lp.PState == "N") && (s == "L") {
		lp.PState = "L1"
	} else if (lp.PState == "L1") && IsCapLetter(s) {
		lp.LambdaBinding += s
	} else if IsInteger(s) && ((lp.PState == "L2") || (lp.PState == "L1")) {
		lp.LambdaBinding += s
		lp.PState = "L2"
	} else if (lp.PState == "L2") && (s == ".") {
		lp.PState = "L3"
	} else if (lp.PState == "L3") && (s == "(") {
		lp.PState = "P"
	} else {
		state := lp.PState + " Lambda Binding"
		l_err = StateCharError(state, s)
	}
	return l_err
}

func (lp *LambdaParser) VParse(s string) error {
	var v_err error
	v_err = nil
	if IsCapLetter(s) &&
		((lp.PState == "N") || (lp.PState == "V1")) {
		lp.CollectionStr += s
		lp.PState = "V1"
	} else if IsInteger(s) &&
		((lp.PState == "V1") || (lp.PState == "V2")) {
		lp.CollectionStr += s
		lp.PState = "V2"
	} else if (lp.PState == "V2") && IsCapLetter(s) {
		lvar := LVar{lp.CollectionStr}
		lp.LExprArr = append(lp.LExprArr, lvar)
		lp.CollectionStr = ""
		lp.CollectionStr += s
		lp.PState = "V1"
	} else if (lp.PState == "V2") && (s == "L") {
		lvar := LVar{lp.CollectionStr}
		lp.LExprArr = append(lp.LExprArr, lvar)
		lp.CollectionStr = ""
		lp.PState = "L1"
	} else if (lp.PState == "V2") && (s == "(") {
		lvar := LVar{lp.CollectionStr}
		lp.LExprArr = append(lp.LExprArr, lvar)
		lp.CollectionStr = ""
		lp.PState = "P"
	} else {
		state := lp.PState + " Variable Reading"
		v_err = StateCharError(state, s)
	}
	return v_err
}

func StateCharError(state, next_str string) error {
	error_template := "Parser in state (%s) encountered invalid next character %s"
	return errors.New(fmt.Sprintf(error_template, state, next_str))
}

func (lp *LambdaParser) DriveParse() ([]LExpr, error) {
	for _, next_str_o := range lp.SrcStr {
		next_str := string(next_str_o)
		major_state := string(lp.PState[0])
		lp.ParenthesesTracker.Update(next_str)
		var step_err error
		step_err = nil
		if (major_state == "L") ||
			((major_state == "N") && (next_str == "L")) {
			step_err = lp.LParse(next_str)
		} else if (major_state == "P") ||
			((major_state == "N") && (next_str == "(")) {
			step_err = lp.PParse(next_str)
		} else if (major_state == "V") ||
			((major_state == "N") && IsCapLetter(next_str)) {
			step_err = lp.VParse(next_str)
		} else {
			state := ""
			if major_state == "L" {
				state = "L Lambda Binding"
			} else if major_state == "P" {
				state = "P Parentheses-Enclosed Expr"
			} else if major_state == "V" {
				state = "V Reading Variable"
			} else {
				state = "N Neutral State"
			}
			step_err = StateCharError(state, next_str)
		}
		if step_err != nil {
			return nil, step_err
		}
	}
	return lp.LExprArr, nil
}

// Main function to convert a given input string into a lambda expression
// Beforehand - uppercase everything, convert L to lambda
//func ParseToLExpr(lstr string) []LExpr {
//	var parser_state string
//	parser_state = "N" // Neutral, L Lambda, V Variable, or P Parentheses
//	track_nest := ParenTracker{Counter: 0}
//	lambda_var := ""
//	content_str := ""
//	var expr_array []LExpr
//	expr_array = nil
//
//	for len(lstr) > 0 {
//		next_str := string(lstr[0])
//		lstr = lstr[1:]
//
//		switch parser_state {
//		default:
//			panic("State not a valid N, P, L, or V state")
//		case "N":
//			if next_str == string('L') {
//				parser_state = "L"
//			} else if next_str == string('(') {
//				track_nest.Update(next_str)
//				parser_state = "P"
//			} else if IsCapLetter(next_str) {
//				content_str += next_str
//				parser_state = "V1"
//			} else {
//				fmt.Println("Invalid character found in neutral parser state", next_str)
//				panic("")
//			}
//		case "P":
//			return expr_array
//		case "V1":
//			if IsCapLetter(next_str) {
//				content_str += next_str
//			} else if IsInteger(next_str) {
//				content_str += next_str
//				parser_state = "V2"
//			} else {
//				panic("All variables must be of the form a1, float72, a720, cat0, etc")
//			}
//		case "V2":
//			if
//		case "L":
//			return expr_array
//		}
//	}
//}

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

	nested_1 := LExpression{LVar{}, []LExpr{&bzbyx}, "(" + bzbyx.LPrint() + ")"}
	fmt.Println(nested_1.LPrint())

	nested_2 := LExpression{LVar{}, []LExpr{&nested_1}, "(" + nested_1.LPrint() + ")"}
	fmt.Println(nested_2.LPrint())

	nested_3 := ConcatenateLExprs([]LExpr{&nested_2})
	fmt.Println(nested_3.LPrint())

	fmt.Println("Testing nested equality: ", nested_3.LEquals(&bzbyx))
}
