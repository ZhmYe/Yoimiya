// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package cs

import (
	"S-gnark/constraint"
	"S-gnark/internal/utils"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// CoeffTable ensure we store unique coefficients in the constraint system
type CoeffTable struct {
	Coefficients []fr.Element
	mCoeffs      map[fr.Element]uint32 // maps coefficient to coeffID
}

func newCoeffTable(capacity int) CoeffTable {
	r := CoeffTable{
		Coefficients: make([]fr.Element, 5, 5+capacity),
		mCoeffs:      make(map[fr.Element]uint32, capacity),
	}

	r.Coefficients[constraint.CoeffIdZero].SetUint64(0)
	r.Coefficients[constraint.CoeffIdOne].SetOne()
	r.Coefficients[constraint.CoeffIdTwo].SetUint64(2)
	r.Coefficients[constraint.CoeffIdMinusOne].SetInt64(-1)
	r.Coefficients[constraint.CoeffIdMinusTwo].SetInt64(-2)

	return r

}

func (ct *CoeffTable) AddCoeff(coeff constraint.Element) uint32 {
	c := (*fr.Element)(coeff[:])
	var cID uint32
	if c.IsZero() {
		cID = constraint.CoeffIdZero
	} else if c.IsOne() {
		cID = constraint.CoeffIdOne
	} else if c.Equal(&two) {
		cID = constraint.CoeffIdTwo
	} else if c.Equal(&minusOne) {
		cID = constraint.CoeffIdMinusOne
	} else if c.Equal(&minusTwo) {
		cID = constraint.CoeffIdMinusTwo
	} else {
		cc := *c
		if id, ok := ct.mCoeffs[cc]; ok {
			cID = id
		} else {
			cID = uint32(len(ct.Coefficients))
			ct.Coefficients = append(ct.Coefficients, cc)
			ct.mCoeffs[cc] = cID
		}
	}
	return cID
}

func (ct *CoeffTable) MakeTerm(coeff constraint.Element, variableID int) constraint.Term {
	cID := ct.AddCoeff(coeff)
	return constraint.Term{VID: uint32(variableID), CID: cID}
}

// CoeffToString implements constraint.Resolver
func (ct *CoeffTable) CoeffToString(cID int) string {
	return ct.Coefficients[cID].String()
}

// implements constraint.Field
type field struct{}

var _ constraint.Field = &field{}

var (
	two      fr.Element
	minusOne fr.Element
	minusTwo fr.Element
)

func init() {
	minusOne.SetOne()
	minusOne.Neg(&minusOne)
	two.SetOne()
	two.Double(&two)
	minusTwo.Neg(&two)
}

func (engine *field) FromInterface(i interface{}) constraint.Element {
	var e fr.Element
	if _, err := e.SetInterface(i); err != nil {
		// need to clean that --> some code path are dissimilar
		// for example setting a fr.Element from an fp.Element
		// fails with the above but succeeds through big int... (2-chains)
		b := utils.FromInterface(i)
		e.SetBigInt(&b)
	}
	var r constraint.Element
	copy(r[:], e[:])
	return r
}
func (engine *field) ToBigInt(c constraint.Element) *big.Int {
	e := (*fr.Element)(c[:])
	r := new(big.Int)
	e.BigInt(r)
	return r

}
func (engine *field) Mul(a, b constraint.Element) constraint.Element {
	_a := (*fr.Element)(a[:])
	_b := (*fr.Element)(b[:])
	_a.Mul(_a, _b)
	return a
}

func (engine *field) Add(a, b constraint.Element) constraint.Element {
	_a := (*fr.Element)(a[:])
	_b := (*fr.Element)(b[:])
	_a.Add(_a, _b)
	return a
}
func (engine *field) Sub(a, b constraint.Element) constraint.Element {
	_a := (*fr.Element)(a[:])
	_b := (*fr.Element)(b[:])
	_a.Sub(_a, _b)
	return a
}
func (engine *field) Neg(a constraint.Element) constraint.Element {
	e := (*fr.Element)(a[:])
	e.Neg(e)
	return a

}
func (engine *field) Inverse(a constraint.Element) (constraint.Element, bool) {
	if a.IsZero() {
		return a, false
	}
	e := (*fr.Element)(a[:])
	if e.IsZero() {
		return a, false
	} else if e.IsOne() {
		return a, true
	}
	var t fr.Element
	t.Neg(e)
	if t.IsOne() {
		return a, true
	}

	e.Inverse(e)
	return a, true
}

func (engine *field) IsOne(a constraint.Element) bool {
	e := (*fr.Element)(a[:])
	return e.IsOne()
}

func (engine *field) One() constraint.Element {
	e := fr.One()
	var r constraint.Element
	copy(r[:], e[:])
	return r
}

func (engine *field) String(a constraint.Element) string {
	e := (*fr.Element)(a[:])
	return e.String()
}

func (engine *field) Uint64(a constraint.Element) (uint64, bool) {
	e := (*fr.Element)(a[:])
	if !e.IsUint64() {
		return 0, false
	}
	return e.Uint64(), true
}
