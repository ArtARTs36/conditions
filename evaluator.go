package conditions

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
)

var (
	falseExpr      = &BooleanLiteral{Val: false}
	defaultEpsilon = float64(1e-6)
)

// SetDefaultEpsilon sets the defaultEpsilon
func SetDefaultEpsilon(ep float64) {
	defaultEpsilon = ep
}

// Evaluate takes an expr and evaluates it using given args
func Evaluate(expr Expr, args map[string]interface{}) (bool, error) {
	return EvaluateWithArgResolver(expr, NewMapArgResolver(args))
}

// EvaluateWithArgResolver takes an expr and evaluates it using given arg resolver
func EvaluateWithArgResolver(expr Expr, args ArgResolver) (bool, error) {
	if expr == nil {
		return false, fmt.Errorf("provided expression is nil")
	}

	result, err := evaluateSubtree(expr, args)
	if err != nil {
		return false, err
	}
	switch n := result.(type) {
	case *BooleanLiteral:
		return n.Val, nil
	}
	return false, fmt.Errorf("unexpected result of the root expression: %#v", result)
}

// evaluateSubtree performs given expr evaluation recursively
func evaluateSubtree(expr Expr, args ArgResolver) (Expr, error) {
	if expr == nil {
		return falseExpr, fmt.Errorf("Provided expression is nil")
	}

	var (
		err    error
		lv, rv Expr
	)

	switch n := expr.(type) {
	case *ParenExpr:
		return evaluateSubtree(n.Expr, args)
	case *BinaryExpr:
		lv, err = evaluateSubtree(n.LHS, args)
		if err != nil {
			return falseExpr, err
		}
		rv, err = evaluateSubtree(n.RHS, args)
		if err != nil {
			return falseExpr, err
		}
		return applyOperator(n.Op, lv, rv)
	case *VarRef:
		//index, err := strconv.Atoi(strings.Replace(n.Val, "$", "", -1))
		index := n.Val
		if err != nil {
			return falseExpr, fmt.Errorf("Failed to resolve argument index %s: %s", n.Val, err.Error())
		}

		arg, err := args.Resolve(index)

		if err != nil {
			return falseExpr, fmt.Errorf("argument %v not resolved: %s", index, err.Error())
		}

		typeof := reflect.TypeOf(arg)
		if typeof == nil {
			return falseExpr, fmt.Errorf("Unsupported argument nil type")
		}

		kind := typeof.Kind()
		switch kind {
		case reflect.Int:
			return &NumberLiteral{Val: float64(arg.(int))}, nil
		case reflect.Int32:
			return &NumberLiteral{Val: float64(arg.(int32))}, nil
		case reflect.Int64:
			return &NumberLiteral{Val: float64(arg.(int64))}, nil
		case reflect.Float32:
			return &NumberLiteral{Val: float64(arg.(float32))}, nil
		case reflect.Float64:
			return &NumberLiteral{Val: float64(arg.(float64))}, nil
		case reflect.String:
			if num, ok := arg.(json.Number); ok {
				f, err := num.Float64()
				if err != nil {
					return falseExpr, fmt.Errorf("Unsupported JSON Number %v type: %s", arg, kind)
				}
				return &NumberLiteral{Val: f}, nil
			}
			return &StringLiteral{Val: arg.(string)}, nil
		case reflect.Bool:
			return &BooleanLiteral{Val: arg.(bool)}, nil
		case reflect.Slice:
			switch arg.(type) {
			case []string:
				ssl := NewSliceStringLiteral(arg.([]string))
				return ssl, nil
			case []int:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]int) {
					snl.Val = append(snl.Val, float64(v))
				}
				return snl, nil
			case []int32:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]int32) {
					snl.Val = append(snl.Val, float64(v))
				}
				return snl, nil
			case []int64:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]int64) {
					snl.Val = append(snl.Val, float64(v))
				}
				return snl, nil
			case []float32:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]float32) {
					snl.Val = append(snl.Val, float64(v))
				}
				return snl, nil
			case []float64:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]float64) {
					snl.Val = append(snl.Val, float64(v))
				}
				return snl, nil
			case []json.Number:
				snl := &SliceNumberLiteral{}
				for _, v := range arg.([]json.Number) {
					f, _ := v.Float64()
					snl.Val = append(snl.Val, f)
				}
				return snl, nil
			case []interface{}:
				items := arg.([]interface{})
				if len(items) != 0 {
					item := items[0]
					switch item.(type) {
					case string:
						val := []string{}
						for _, v := range items {
							val = append(val, v.(string))
						}
						ssl := NewSliceStringLiteral(val)
						return ssl, nil
					case float64:
						snl := &SliceNumberLiteral{}
						for _, v := range items {
							snl.Val = append(snl.Val, v.(float64))
						}
						return snl, nil
					case json.Number:
						snl := &SliceNumberLiteral{}
						for _, v := range items {
							f, _ := v.(json.Number).Float64()
							snl.Val = append(snl.Val, f)
						}
						return snl, nil
					}
				}
			}
		case reflect.Struct:
			return createCollectionLiteral(n.Val, kind, arg)
		case reflect.Ptr:
			return createCollectionLiteral(n.Val, kind, arg)
		}

		return falseExpr, fmt.Errorf("Unsupported argument %s type: %s", n.Val, kind)
	}

	return expr, nil
}

func createCollectionLiteral(argName string, itemType reflect.Kind, arg interface{}) (Expr, error) {
	numCollection := tryCreateNumberCollectionLiteral(arg)

	if numCollection != nil {
		return numCollection, nil
	}

	strCollection := tryCreateStringCollectionLiteral(arg)

	if strCollection != nil {
		return strCollection, nil
	}

	return falseExpr, fmt.Errorf("unsupported structure of argument %s type: %s", argName, itemType)
}

func tryCreateNumberCollectionLiteral(arg interface{}) *NumberCollectionLiteral {
	numCollection, isNumCollection := arg.(MapNumberCollection)

	if isNumCollection {
		return &NumberCollectionLiteral{Val: &numCollection}
	}

	numCollectionRef, isNumCollectionRef := arg.(*MapNumberCollection)

	if isNumCollectionRef {
		return &NumberCollectionLiteral{Val: numCollectionRef}
	}

	return nil
}

func tryCreateStringCollectionLiteral(arg interface{}) *StringCollectionLiteral {
	numCollection, isNumCollection := arg.(MapStringCollection)

	if isNumCollection {
		return &StringCollectionLiteral{Val: &numCollection}
	}

	numCollectionRef, isNumCollectionRef := arg.(*MapStringCollection)

	if isNumCollectionRef {
		return &StringCollectionLiteral{Val: numCollectionRef}
	}

	return nil
}

// applyOperator is a dispatcher of the evaluation according to operator
func applyOperator(op Token, l, r Expr) (*BooleanLiteral, error) {
	switch op {
	case AND:
		return applyAND(l, r)
	case OR:
		return applyOR(l, r)
	case EQ:
		return applyEQ(l, r)
	case NEQ:
		return applyNQ(l, r)
	case GT:
		return applyGT(l, r)
	case GTE:
		return applyGTE(l, r)
	case LT:
		return applyLT(l, r)
	case LTE:
		return applyLTE(l, r)
	case XOR:
		return applyXOR(l, r)
	case NAND:
		return applyNAND(l, r)
	case IN:
		return applyIN(l, r)
	case NOTIN:
		return applyNOTIN(l, r)
	case EREG:
		return applyEREG(l, r)
	case NEREG:
		return applyNEREG(l, r)
	case CONTAINS:
		return applyCONTAINS(l, r)
	case NOTCONTAINS:
		return applyNOTCONTAINS(l, r)
	}
	return &BooleanLiteral{Val: false}, fmt.Errorf("Unsupported operator: %s", op)
}

// applyEREG applies EREG operation to l/r operands
func applyNEREG(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyEREG(l, r)
	result.Val = !result.Val
	return result, err
}

// applyEREG applies EREG operation to l/r operands
func applyEREG(l, r Expr) (*BooleanLiteral, error) {
	var (
		a     string
		b     string
		err   error
		match bool
	)
	a, err = getString(l)
	if err != nil {
		return nil, err
	}

	b, err = getString(r)
	if err != nil {
		return nil, err
	}
	match = false
	match, err = regexp.MatchString(b, a)

	// pp.Print(a, b, match)
	return &BooleanLiteral{Val: match}, err
}

// applyNOTIN applies NOT IN operation to l/r operands
func applyNOTIN(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyIN(l, r)
	if err != nil {
		return nil, err
	}
	result.Val = !result.Val
	return result, err
}

// applyIN applies IN operation to l/r operands
func applyIN(l, r Expr) (*BooleanLiteral, error) {
	var (
		err   error
		found bool
	)
	// pp.Print(l)
	switch t := l.(type) {
	case *StringLiteral:
		var a string
		var b map[string]struct{}

		a, err = getString(l)
		if err != nil {
			return nil, err
		}

		switch r.(type) {
		case *StringCollectionLiteral:
			rNumbers := r.(*StringCollectionLiteral)

			return &BooleanLiteral{Val: rNumbers.Val.Has(a)}, nil
		}

		b, err = getMapString(r)
		if err != nil {
			return nil, err
		}
		_, found = b[a]
	case *NumberLiteral:
		var a float64
		var b []float64
		a, err = getNumber(l)
		if err != nil {
			return nil, err
		}

		switch r.(type) {
		case *NumberCollectionLiteral:
			rNumbers := r.(*NumberCollectionLiteral)

			return &BooleanLiteral{Val: rNumbers.Val.Has(a)}, nil
		}

		b, err = getSliceNumber(r)

		if err != nil {
			return nil, err
		}

		found = false
		for _, e := range b {
			if float64Equal(a, e, defaultEpsilon) {
				found = true

				break
			}
		}
	default:
		return nil, fmt.Errorf("Can not evaluate Literal of unknow type %s %T", t, t)
	}

	return &BooleanLiteral{Val: found}, nil
}

// applyCONTAINS applies CONTAINS operation to l/r operands
func applyCONTAINS(l, r Expr) (*BooleanLiteral, error) {
	return applyIN(r, l)
}

// applyNOTCONTAINS applies NOT CONTAINS operation to l/r operands
func applyNOTCONTAINS(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyCONTAINS(l, r)
	if err != nil {
		return nil, err
	}
	result.Val = !result.Val
	return result, err
}

// applyXOR applies || operation to l/r operands
func applyXOR(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a != b)}, nil
}

// applyNAND applies NAND operation to l/r operands
func applyNAND(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (!(a && b))}, nil
}

// applyAND applies && operation to l/r operands
func applyAND(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a && b)}, nil
}

// applyOR applies || operation to l/r operands
func applyOR(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b bool
		err  error
	)
	a, err = getBoolean(l)
	if err != nil {
		return nil, err
	}
	b, err = getBoolean(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a || b)}, nil
}

// applyEQ applies == operation to l/r operands
func applyEQ(l, r Expr) (*BooleanLiteral, error) {
	var (
		as, bs string
		an, bn float64
		ab, bb bool
		err    error
	)
	as, err = getString(l)
	if err == nil {
		bs, err = getString(r)
		if err != nil {
			return falseExpr, fmt.Errorf("cannot compare string(%v) with non-string(%v)", l, r)
		}
		return &BooleanLiteral{Val: (as == bs)}, nil
	}
	an, err = getNumber(l)
	if err == nil {
		bn, err = getNumber(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare number with non-number")
		}
		return &BooleanLiteral{Val: float64Equal(an, bn, defaultEpsilon)}, nil
	}
	ab, err = getBoolean(l)
	if err == nil {
		bb, err = getBoolean(r)
		if err != nil {
			return falseExpr, fmt.Errorf("Cannot compare boolean with non-boolean")
		}
		return &BooleanLiteral{Val: (ab == bb)}, nil
	}
	return falseExpr, nil
}

// applyNQ applies != operation to l/r operands
func applyNQ(l, r Expr) (*BooleanLiteral, error) {
	result, err := applyEQ(l, r)
	if err != nil {
		return nil, err
	}
	result.Val = !result.Val
	return result, err
}

// applyGT applies > operation to l/r operands
func applyGT(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a > b)}, nil
}

// applyGTE applies >= operation to l/r operands
func applyGTE(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a > b) || float64Equal(a, b, defaultEpsilon)}, nil
}

// applyLT applies < operation to l/r operands
func applyLT(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return nil, err
	}
	b, err = getNumber(r)
	if err != nil {
		return nil, err
	}
	return &BooleanLiteral{Val: (a < b)}, nil
}

// applyLTE applies <= operation to l/r operands
func applyLTE(l, r Expr) (*BooleanLiteral, error) {
	var (
		a, b float64
		err  error
	)
	a, err = getNumber(l)
	if err != nil {
		return falseExpr, err
	}
	b, err = getNumber(r)
	if err != nil {
		return falseExpr, err
	}
	return &BooleanLiteral{Val: (a < b) || float64Equal(a, b, defaultEpsilon)}, nil
}

// getBoolean performs type assertion and returns boolean value or error
func getBoolean(e Expr) (bool, error) {
	switch n := e.(type) {
	case *BooleanLiteral:
		return n.Val, nil
	default:
		return false, fmt.Errorf("Literal is not a boolean: %v", n)
	}
}

// getString performs type assertion and returns string value or error
func getString(e Expr) (string, error) {
	switch n := e.(type) {
	case *StringLiteral:
		return n.Val, nil
	default:
		return "", fmt.Errorf("Literal is not a string: %v", n)
	}
}

// getSliceNumber performs type assertion and returns []float64 value or error
func getSliceNumber(e Expr) ([]float64, error) {
	switch n := e.(type) {
	case *SliceNumberLiteral:
		return n.Val, nil
	default:
		return []float64{}, fmt.Errorf("Literal is not a slice of float64: %v", n)
	}
}

// getMapString performs type assertion and returns map[string]struct{} value or error
func getMapString(e Expr) (map[string]struct{}, error) {
	switch n := e.(type) {
	case *SliceStringLiteral:
		return n.m, nil
	default:
		return nil, fmt.Errorf("Literal is not a slice of string: %v", n)
	}
}

// getNumber performs type assertion and returns float64 value or error
func getNumber(e Expr) (float64, error) {
	switch n := e.(type) {
	case *NumberLiteral:
		return n.Val, nil
	default:
		return 0, fmt.Errorf("Literal is not a number: %v", n)
	}
}

func float64Equal(a float64, b float64, epsilon float64) bool {
	absA := math.Abs(a)
	absB := math.Abs(b)
	diff := math.Abs(a - b)
	zero := float64(0)

	if a == b {
		return true
	}
	if diff > epsilon {
		return false
	}
	if a == zero || b == zero {
		return diff < epsilon*math.SmallestNonzeroFloat32
	}
	return diff/math.Min((absA+absB), math.MaxFloat64) < epsilon
}
