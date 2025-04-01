package main

// import (
// 	"fmt"
// )

type ParserState int

const (
	_   ParserState = iota
	I_i             // Initial state
	L_i             // Read L lambda expressoin start
	LV1             // Captured letter for a var for a variable binding
	LV2             // Captured number for a var for a variable binding\
	LV3             // Read "." ending a variable binding.
	LP1             // Read "(" at beginning of function body, captures any further input read
	L_f             // Read corresponding closing ")" and process captured strings.
	// TERMINAL for creating complete FUNCTION expression
	V_i // Captured letter for a var (not part of a variable binding of a func)
	V_f // Captured number for a var (not part of a variable binding of a func)
	// TERMINAL for creating complete VAR expression.
	P_i // Read "(" at beginning of a parenthesized L-expr (not immediately bound by a variable).
	// Captures any further input read
	P_f // Read corresponding closing ")" and process captured strings.
	// TERMINAL for creating complete CONCAT expressions.

	E0    // End State. Should always succeed some neutral/TERMINAL state.
	DUMMY // Represents arbitrary state
)

type Transition struct {
	S_i ParserState
	S_f ParserState
}

// type Tuple [T any, G any] struct {
// 	Elt1 T
// 	Elt2 G
// }

// type TransitionInput Tuple[ParserState, any]

type Parser struct {
	Exprs           []LExpr
	Parentheses     ParenTracker
	FnArg           string
	CapturedStrings string
	TState          Transition
}

func Parser_Init() Parser {
	return Parser{
		Exprs:           nil,
		Parentheses:     ParenTracker{},
		FnArg:           "",
		CapturedStrings: "",
		TState:          Transition{S_i: DUMMY, S_f: I_i},
	}
}

type TransitionCallback func(s string, p Parser) Parser
type ParserValidator func(p Parser)
type TransitionMapper func(p Parser, s string) (Transition, error)

type TransitionExecutor struct {
	TransitionCallbackMap map[Transition][]TransitionCallback
	TransitionMap         map[ParserState]TransitionMapper
}

func (t *TransitionExecutor) Parse(target_str string) (Parser, error) {
	p := Parser_Init()

	for _, char := range target_str {
		current_transition, transition_err := t.TransitionMap[p.TState.S_f](p, string(char))
		if transition_err != nil {
			return p, transition_err
		}
		p.TState = current_transition
		p.Parentheses.Update(string(char))
		p = t.Apply(string(char), p)
	}
	// Transition into terminal state E0 and run final callbacks in response
	final_transition := Transition{S_i: p.TState.S_f, S_f: E0}
	p.TState = final_transition
	p = t.Apply("", p)
	return p, nil
}

func (t *TransitionExecutor) FilterCallbacks(ts Transition) []TransitionCallback {
	ts_filter := func(ts_input Transition) bool {
		exact_ts_match := ts == ts_input
		initial_only := Transition{S_i: ts.S_i, S_f: DUMMY} == ts_input
		final_only := Transition{S_i: DUMMY, S_f: ts.S_f} == ts_input
		return exact_ts_match || initial_only || final_only
	}
	selected_callbacks := []TransitionCallback{}
	for k, v := range t.TransitionCallbackMap {
		if ts_filter(k) {
			selected_callbacks = append(selected_callbacks, v...)
		}
	}
	return selected_callbacks
}

func (t *TransitionExecutor) Apply(s string, p Parser) Parser {
	callback_list := t.FilterCallbacks((p.TState))
	for _, callback := range callback_list {
		p = callback(s, p)
	}
	return p
}

func CaptureFnArg(s string, p Parser) Parser {
	p.FnArg += s
	return p
}

func CaptureStrings(s string, p Parser) Parser {
	p.CapturedStrings += s
	return p
}
