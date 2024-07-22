package Circuit4Neg

import (
	"Yoimiya/frontend"
)

type NegCircuit struct {
	X frontend.Variable `gnark:",public"` // a_1,a_2
}

func (c *NegCircuit) Define(api frontend.API) error {
	//api.ToBinary()
	myCmp := func(i1 frontend.Variable, i2 frontend.Variable) frontend.Variable {
		nbBits := 254
		// Convert to binary representation, assuming the highest bit is the sign bit
		bi1 := api.ToBinary(i1)
		bi2 := api.ToBinary(i2)
		//fmt.Println(bi1, bi2)
		// Sign bits
		sign1 := bi1[nbBits-1]
		sign2 := bi2[nbBits-1]

		// Compare signs first
		isNegative1 := api.IsZero(sign1) // 如果是负数则为1
		isNegative2 := api.IsZero(sign2)

		// If signs are different, determine the result based on sign
		// If i1 is negative and i2 is positive, i1 < i2
		// If i1 is positive and i2 is negative, i1 > i2
		signComparison := api.Xor(isNegative1, isNegative2)

		//signComparison = builder.Select(builder.And(isNegative2, builder.Select(isNegative1, 0, 1)), 1, signComparison)

		// If signs are the same, compare the absolute values
		res := frontend.Variable(0)
		for i := nbBits - 2; i >= 0; i-- { // nbBits-2 to skip the sign bit
			iszeroi1 := api.IsZero(bi1[i])
			iszeroi2 := api.IsZero(bi2[i])

			i1i2 := api.And(bi1[i], iszeroi2)
			i2i1 := api.And(bi2[i], iszeroi1)

			n := api.Select(i2i1, -1, 0)
			m := api.Select(i1i2, 1, n)

			res = api.Select(api.IsZero(res), m, res)
		}
		return api.Select(signComparison, api.Select(isNegative2, -1, 1), api.Select(isNegative1, res, api.Neg(res)))
	}
	api.AssertIsEqual(myCmp(0, 0), 1)
	return nil
}
