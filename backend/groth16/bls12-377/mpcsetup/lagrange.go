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

package mpcsetup

import (
	"math/big"
	"math/bits"
	"runtime"

	"S-gnark/internal/utils"
	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
)

// TODO use gnark-crypto for this op
func lagrangeCoeffsG1(powers []curve.G1Affine, size int) []curve.G1Affine {
	coeffs := make([]curve.G1Affine, size)
	copy(coeffs, powers[:size])
	domain := fft.NewDomain(uint64(size))
	numCPU := uint64(runtime.NumCPU())
	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(numCPU))

	difFFTG1(coeffs, domain.TwiddlesInv, 0, maxSplits, nil)
	bitReverse(coeffs)

	var invBigint big.Int
	domain.CardinalityInv.BigInt(&invBigint)

	utils.Parallelize(size, func(start, end int) {
		for i := start; i < end; i++ {
			coeffs[i].ScalarMultiplication(&coeffs[i], &invBigint)
		}
	})
	return coeffs
}

// TODO use gnark-crypto for this op
func lagrangeCoeffsG2(powers []curve.G2Affine, size int) []curve.G2Affine {
	coeffs := make([]curve.G2Affine, size)
	copy(coeffs, powers[:size])
	domain := fft.NewDomain(uint64(size))
	numCPU := uint64(runtime.NumCPU())
	maxSplits := bits.TrailingZeros64(ecc.NextPowerOfTwo(numCPU))

	difFFTG2(coeffs, domain.TwiddlesInv, 0, maxSplits, nil)
	bitReverse(coeffs)

	var invBigint big.Int
	domain.CardinalityInv.BigInt(&invBigint)

	utils.Parallelize(size, func(start, end int) {
		for i := start; i < end; i++ {
			coeffs[i].ScalarMultiplication(&coeffs[i], &invBigint)
		}
	})
	return coeffs
}

func butterflyG1(a *curve.G1Affine, b *curve.G1Affine) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

func butterflyG2(a *curve.G2Affine, b *curve.G2Affine) {
	t := *a
	a.Add(a, b)
	b.Sub(&t, b)
}

// kerDIF8 is a kernel that process a FFT of size 8
func kerDIF8G1(a []curve.G1Affine, twiddles [][]fr.Element, stage int) {
	butterflyG1(&a[0], &a[4])
	butterflyG1(&a[1], &a[5])
	butterflyG1(&a[2], &a[6])
	butterflyG1(&a[3], &a[7])

	var twiddle big.Int
	twiddles[stage+0][1].BigInt(&twiddle)
	a[5].ScalarMultiplication(&a[5], &twiddle)
	twiddles[stage+0][2].BigInt(&twiddle)
	a[6].ScalarMultiplication(&a[6], &twiddle)
	twiddles[stage+0][3].BigInt(&twiddle)
	a[7].ScalarMultiplication(&a[7], &twiddle)
	butterflyG1(&a[0], &a[2])
	butterflyG1(&a[1], &a[3])
	butterflyG1(&a[4], &a[6])
	butterflyG1(&a[5], &a[7])
	twiddles[stage+1][1].BigInt(&twiddle)
	a[3].ScalarMultiplication(&a[3], &twiddle)
	twiddles[stage+1][1].BigInt(&twiddle)
	a[7].ScalarMultiplication(&a[7], &twiddle)
	butterflyG1(&a[0], &a[1])
	butterflyG1(&a[2], &a[3])
	butterflyG1(&a[4], &a[5])
	butterflyG1(&a[6], &a[7])
}

// kerDIF8 is a kernel that process a FFT of size 8
func kerDIF8G2(a []curve.G2Affine, twiddles [][]fr.Element, stage int) {
	butterflyG2(&a[0], &a[4])
	butterflyG2(&a[1], &a[5])
	butterflyG2(&a[2], &a[6])
	butterflyG2(&a[3], &a[7])

	var twiddle big.Int
	twiddles[stage+0][1].BigInt(&twiddle)
	a[5].ScalarMultiplication(&a[5], &twiddle)
	twiddles[stage+0][2].BigInt(&twiddle)
	a[6].ScalarMultiplication(&a[6], &twiddle)
	twiddles[stage+0][3].BigInt(&twiddle)
	a[7].ScalarMultiplication(&a[7], &twiddle)
	butterflyG2(&a[0], &a[2])
	butterflyG2(&a[1], &a[3])
	butterflyG2(&a[4], &a[6])
	butterflyG2(&a[5], &a[7])
	twiddles[stage+1][1].BigInt(&twiddle)
	a[3].ScalarMultiplication(&a[3], &twiddle)
	twiddles[stage+1][1].BigInt(&twiddle)
	a[7].ScalarMultiplication(&a[7], &twiddle)
	butterflyG2(&a[0], &a[1])
	butterflyG2(&a[2], &a[3])
	butterflyG2(&a[4], &a[5])
	butterflyG2(&a[6], &a[7])
}

func difFFTG1(a []curve.G1Affine, twiddles [][]fr.Element, stage, maxSplits int, chDone chan struct{}) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if n == 8 {
		kerDIF8G1(a, twiddles, stage)
		return
	}
	m := n >> 1

	butterflyG1(&a[0], &a[m])

	var twiddle big.Int
	for i := 1; i < m; i++ {
		butterflyG1(&a[i], &a[i+m])
		twiddles[stage][i].BigInt(&twiddle)
		a[i+m].ScalarMultiplication(&a[i+m], &twiddle)
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go difFFTG1(a[m:n], twiddles, nextStage, maxSplits, chDone)
		difFFTG1(a[0:m], twiddles, nextStage, maxSplits, nil)
		<-chDone
	} else {
		difFFTG1(a[0:m], twiddles, nextStage, maxSplits, nil)
		difFFTG1(a[m:n], twiddles, nextStage, maxSplits, nil)
	}
}
func difFFTG2(a []curve.G2Affine, twiddles [][]fr.Element, stage, maxSplits int, chDone chan struct{}) {
	if chDone != nil {
		defer close(chDone)
	}

	n := len(a)
	if n == 1 {
		return
	} else if n == 8 {
		kerDIF8G2(a, twiddles, stage)
		return
	}
	m := n >> 1

	butterflyG2(&a[0], &a[m])

	var twiddle big.Int
	for i := 1; i < m; i++ {
		butterflyG2(&a[i], &a[i+m])
		twiddles[stage][i].BigInt(&twiddle)
		a[i+m].ScalarMultiplication(&a[i+m], &twiddle)
	}

	if m == 1 {
		return
	}

	nextStage := stage + 1
	if stage < maxSplits {
		chDone := make(chan struct{}, 1)
		go difFFTG2(a[m:n], twiddles, nextStage, maxSplits, chDone)
		difFFTG2(a[0:m], twiddles, nextStage, maxSplits, nil)
		<-chDone
	} else {
		difFFTG2(a[0:m], twiddles, nextStage, maxSplits, nil)
		difFFTG2(a[m:n], twiddles, nextStage, maxSplits, nil)
	}
}
