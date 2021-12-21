package vm

import (
	"log"
	"testing"

	"github.com/npillmayer/arithm/polyn"
)

type X struct { // helper for quick construction of polynomials
	i int
	c float64
}

func polynom(c float64, tms ...X) polyn.Polynomial { // construct a polynomial
	p := polyn.NewConstantPolynomial(c)
	for _, t := range tms {
		p.SetTerm(t.i, t.c)
	}
	return p
}

func TestStackCreate(t *testing.T) {
	p := polynom(5, X{1, 1}, X{2, 2})
	log.Printf("p = %s\n", p.String())
}

/*
func TestStackVar1(t *testing.T) {
	est := NewExprStack()
	est.AnnounceVariable(runtime.NewTag("sym1"))
	p := polynom(5, X{1, 1}, X{2, 2})
	log.Printf("p = %s\n", p.TraceString(est))
}
*/

func TestStackVar2(t *testing.T) {
	est := NewExprStack()
	est.PushConstant(4711)
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
}

/*
func TestStackVar3(t *testing.T) {
	est := NewExprStack()
	v := runtime.NewTag("a")
	est.PushVariable(v, nil)
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
}
*/

/*
func TestStackAdd(t *testing.T) {
	est := NewExprStack()
	v1 := runtime.NewTag("a")
	est.PushVariable(v1, nil)
	v2 := runtime.NewTag("b")
	est.PushVariable(v2, nil)
	est.AddTOS2OS()
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
	if est.Top().XPolyn.GetConstantValue() != 0 {
		t.Fail()
	}
	if est.Top().XPolyn.Terms.Size() != 3 {
		t.Fail()
	}
}
*/

/*
func TestStackSubtract(t *testing.T) {
	est := NewExprStack()
	v1 := runtime.NewTag("a")
	est.PushVariable(v1, nil)
	a := v1.ID
	p := polynom(2.0, X{int(a), 3.0})
	est.Push(NewNumericExpression(p)) // push p = 3a + 2
	est.SubtractTOS2OS()              // should result in p = -2a - 2
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
	if est.Top().XPolyn.GetConstantValue() != -2.0 {
		t.Fail()
	}
	if est.Top().XPolyn.Terms.Size() != 2 {
		t.Fail()
	}
}
*/

/*
func TestStackMultiply(t *testing.T) {
	est := NewExprStack()
	v1 := runtime.NewTag("a")
	est.PushVariable(v1, nil)
	p := polynom(2.0) // constant
	est.Push(NewNumericExpression(p))
	est.MultiplyTOS2OS()
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
	if est.Top().XPolyn.GetConstantValue() != 0.0 {
		t.Fail()
	}
	if est.Top().XPolyn.Terms.Size() != 2 {
		t.Fail()
	}
}
*/

/*
func TestStackDivide(t *testing.T) {
	est := NewExprStack()
	v1 := runtime.NewTag("a")
	est.PushVariable(v1, nil)
	p := polynom(2.0)                 // constant
	est.Push(NewNumericExpression(p)) // push p = 2
	est.DivideTOS2OS()
	log.Printf("TOS = %s\n", est.Top().XPolyn.TraceString(est))
	if est.Top().XPolyn.GetConstantValue() != 0.0 {
		t.Fail()
	}
	if est.Top().XPolyn.Terms.Size() != 2 {
		t.Fail()
	}
}
*/

/*
func TestStackEquation(t *testing.T) {
	est := NewExprStack()
	v1 := runtime.NewTag("a")
	est.PushVariable(v1, nil)
	v2 := runtime.NewTag("b")
	est.PushVariable(v2, nil)
	est.Dump()
	est.EquateTOS2OS()
	if !est.IsEmpty() {
		t.Fail()
	}
	est.leq.Dump(est)
}
*/
