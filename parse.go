package main

import "fmt"

// import (
// 	"fmt"
// )

type ParserState int

const (
	_   ParserState = iota
	I_i             // Initial state
	L_i             // Read "L" lambda expression start
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

// Create 1 map per state to return subsequent state given a certain string

// I_i and all states that complete an expression (other than V_f) share the same mapping.
func End_of_Expression_Mapper(p Parser, s string) (Transition, error) {
	var next_state ParserState
	if IsCapLetter(s) {
		next_state = V_i
	} else if s == "(" {
		next_state = P_i
	} else if s == "L" {
		next_state = L_i
	} else {
		return p.TState, fmt.Errorf(
			"Parsed character (%v) not valid start character for any LExpr"+
				"following state %v. Must be either L or an alphabetical"+
				" character or a ( start parenthesis",
			s,
			p.TState.S_f,
		)
	}
	transition := Transition{S_f: next_state, S_i: p.TState.S_f}
	return transition, nil

}

func I_i_Mapper(p Parser, s string) (Transition, error) {
	return End_of_Expression_Mapper(p, s)
}

func L_f_Mapper(p Parser, s string) (Transition, error) {
	return End_of_Expression_Mapper(p, s)
}

func P_f_Mapper(p Parser, s string) (Transition, error) {
	return End_of_Expression_Mapper(p, s)
}

// Mappers for V states (Binding a Variable term)
func V_i_Mapper(p Parser, s string) (Transition, error) {
	var next_state ParserState
	// TODO: This should be looking for non-L capital letters
	if IsCapLetter(s) {
		next_state = V_i
	} else if IsInteger(s) {
		next_state = V_f
	} else {
		return p.TState, fmt.Errorf(
			"Currently parsing Variable expression and should only encounter "+
				"letters or numbers, encountered char %v",
			s,
		)
	}
	transition := Transition{S_f: next_state, S_i: p.TState.S_f}
	return transition, nil
}

// NOTE: This is a special case of End_of_Expression_Mapper. All other LExpr end on ) char and next character
// begins next LExpr unambiguously. Here, Variables end on a # char but end of Variable can have multilple digits
// i.e. A1(...), A12(...), A123(...) all depict Variables preceding more expressions. End_of_Expression_Mapper
// must be modified here to take into account arbitrarily long int sequence at the end of a Variable expression.
func V_f_Mapper(p Parser, s string) (Transition, error) {
	var next_state ParserState
	if IsInteger(s) {
		next_state = V_f
		transition := Transition{S_f: next_state, S_i: p.TState.S_f}
		return transition, nil
	}
	return End_of_Expression_Mapper(p, s)
}

// Mapper for P state (mapping parentheticals)
func P_i_Mapper(p Parser, s string) (Transition, error) {
	var next_state ParserState
	if (s == ")") && (p.NestTracker.Counter == 0) {
		next_state = P_f
		return Transition{S_f: next_state, S_i: p.TState.S_f}, nil
	}
	next_state = P_i
	return Transition{S_f: next_state, S_i: p.TState.S_f}, nil
}

// Mappers for L states (mapping lambda-bound expression)
func L_i_Mapper(p Parser, s string) (Transition, error) {
	return Transition{}, nil
}

func LV1_Mapper(p Parser, s string) (Transition, error) {
	return Transition{}, nil
}

func LV2_Mapper(p Parser, s string) (Transition, error) {
	return Transition{}, nil
}

func LV3_Mapper(p Parser, s string) (Transition, error) {
	return Transition{}, nil
}

func LP1_Mapper(p Parser, s string) (Transition, error) {
	return Transition{}, nil
}

type Transition struct {
	S_i ParserState
	S_f ParserState
}

type Parser struct {
	Exprs         []LExpr
	NestTracker   ParenTracker
	LVar          string
	Parenthetical string
	TState        Transition
}

func Parser_Init() Parser {
	return Parser{
		Exprs:         nil,
		NestTracker:   ParenTracker{Counter: 0},
		LVar:          "",
		Parenthetical: "",
		TState:        Transition{S_i: DUMMY, S_f: I_i},
	}
}

type TransitionCallback func(p Parser, s string) (Parser, error)
type ParserValidator func(p Parser)
type TransitionMapper func(p Parser, s string) (Transition, error)

type TransitionExecutor struct {
	TransitionCallbackMap map[Transition][]TransitionCallback
	TransitionMap         map[ParserState]TransitionMapper
}

func (t *TransitionExecutor) Parse(target_str string) (Parser, error) {
	p := Parser_Init()

	for _, char := range target_str {
		// Keep track of current parentheses nesting level (specifically for LP1, P_i, and P_f states)
		// NOTE: TransitionMapper has parser and next char as input. p.NestTracker has nesting level
		// INCLUDING NEXT CHAR. So if next char is ) and parse level is 0, then closing ) has been found
		// for a previous opening ( at the same depth. Important for designing P_i and LP1 mappers.
		p.NestTracker.Update(string(char))
		// For current state, find and apply callback to determine next state using next char
		current_transition, transition_err := t.TransitionMap[p.TState.S_f](p, string(char))
		if transition_err != nil {
			// TODO: Nest error in context including most recent state and last read char
			return p, transition_err
		}
		p.TState = current_transition
		var callback_err error = nil
		// Apply all callbacks necessary based off most recent and current states in Parser Transition field
		p, callback_err = t.Apply(p, string(char))
		if callback_err != nil {
			// TODO: Nest error in context including most recent state, last read char, callback, etc
			return p, callback_err
		}
	}
	// Transition into terminal state E0 and run final callbacks in response
	final_transition := Transition{S_i: p.TState.S_f, S_f: E0}
	p.TState = final_transition
	var callback_err error = nil
	p, callback_err = t.Apply(p, "")
	if callback_err != nil {
		return p, callback_err
	}
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

func (t *TransitionExecutor) Apply(p Parser, s string) (Parser, error) {
	callback_list := t.FilterCallbacks((p.TState))
	for _, callback := range callback_list {
		var callback_err error = nil
		p, callback_err = callback(p, s)
		if callback_err != nil {
			// TODO: Nest in more descriptive error
			return p, callback_err
		}
	}
	return p, nil
}

// Effective Calbacks - callbacks which alter the contents of parser
func (t *TransitionExecutor) BuildLVar(p Parser, s string) (Parser, error) {
	p.LVar += s
	return p, nil
}

func (t *TransitionExecutor) BuildParenthetical(p Parser, s string) (Parser, error) {
	p.Parenthetical += s
	return p, nil
}

func (t *TransitionExecutor) CaptureLVar(p Parser, s string) (Parser, error) {
	new_lvar := LVar{Symbol: p.LVar}
	p.LVar = ""
	p.Exprs = append(p.Exprs, &new_lvar)
	return p, nil
}

func (t *TransitionExecutor) CaptureParenthetical(p Parser, s string) (Parser, error) {
	inner_parse, parse_err := t.Parse(p.Parenthetical)
	if parse_err != nil {
		parse_err = fmt.Errorf(
			"Following error emerges while parsing %v within a parenthetical:\n%w",
			p.Parenthetical,
			parse_err,
		)
		return p, parse_err
	}
	new_lexpr := ConcatenateLExprs(inner_parse.Exprs)
	p.Parenthetical = ""
	p.Exprs = append(p.Exprs, &new_lexpr)
	return p, nil
}

func (t *TransitionExecutor) CaptureLambda(p Parser, s string) (Parser, error) {
	inner_parse, parse_err := t.Parse(p.Parenthetical)
	if parse_err != nil {
		parse_err = fmt.Errorf(
			"Following error emerges while parsing %v within a parenthetical:\n%w",
			p.Parenthetical,
			parse_err,
		)
		return p, parse_err
	}
	//TODO: Add sturdier check that p.LVar actually fits var criteria
	// consider adding a Parser method to simultaneously blank out LVar field
	// and return a LVar term
	if len(p.LVar) == 0 {
		return p, fmt.Errorf(
			"Attempting to create lambda expression from %v without having a valid variable to bind with",
			p.Parenthetical,
		)
	}
	new_lexpr := ConcatenateLExprs(inner_parse.Exprs)
	new_lambda := new_lexpr.LAbstract(LVar{Symbol: p.LVar})
	p.Parenthetical = ""
	p.LVar = ""
	p.Exprs = append(p.Exprs, new_lambda)
	return p, nil
}

// TODO: Add validation callbacks which just error out if a certain condition holds on p Parser
