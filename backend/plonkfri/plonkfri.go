// Copyright 2020 ConsenSys AG
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

// Package plonkfri implements PLONK Zero Knowledge Proof system, with FRI as commitment scheme.

package plonkfri

import (
	"Yoimiya/backend"
	"Yoimiya/constraint"

	"Yoimiya/backend/witness"
	cs_bls12377 "Yoimiya/constraint/bls12-377"
	cs_bls12381 "Yoimiya/constraint/bls12-381"
	cs_bls24315 "Yoimiya/constraint/bls24-315"
	cs_bls24317 "Yoimiya/constraint/bls24-317"
	cs_bn254 "Yoimiya/constraint/bn254"
	cs_bw6633 "Yoimiya/constraint/bw6-633"
	cs_bw6761 "Yoimiya/constraint/bw6-761"

	plonk_bls12377 "Yoimiya/backend/plonkfri/bls12-377"
	plonk_bls12381 "Yoimiya/backend/plonkfri/bls12-381"
	plonk_bls24315 "Yoimiya/backend/plonkfri/bls24-315"
	plonk_bls24317 "Yoimiya/backend/plonkfri/bls24-317"
	plonk_bn254 "Yoimiya/backend/plonkfri/bn254"
	plonk_bw6633 "Yoimiya/backend/plonkfri/bw6-633"
	plonk_bw6761 "Yoimiya/backend/plonkfri/bw6-761"

	fr_bls12377 "github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	fr_bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	fr_bls24315 "github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	fr_bls24317 "github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	fr_bn254 "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	fr_bw6633 "github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
	fr_bw6761 "github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// Proof represents a Plonk proof generated by plonk.Prove
//
// it's underlying implementation is curve specific (see gnark/internal/backend)
type Proof interface {
	// io.WriterTo
	// io.ReaderFrom
}

// ProvingKey represents a plonk ProvingKey
//
// it's underlying implementation is strongly typed with the curve (see gnark/internal/backend)
type ProvingKey interface {
	// io.WriterTo
	// io.ReaderFrom
	VerifyingKey() interface{}
}

// VerifyingKey represents a plonk VerifyingKey
//
// it's underlying implementation is strongly typed with the curve (see gnark/internal/backend)
type VerifyingKey interface {
	// io.WriterTo
	// io.ReaderFrom
	// InitKZG(srs kzg.SRS) error
	NbPublicWitness() int // number of elements expected in the public witness
}

// Setup prepares the public data associated to a circuit + public inputs.
func Setup(ccs constraint.ConstraintSystem) (ProvingKey, VerifyingKey, error) {

	switch tccs := ccs.(type) {
	case *cs_bn254.SparseR1CS:
		return plonk_bn254.Setup(tccs)
	case *cs_bls12381.SparseR1CS:
		return plonk_bls12381.Setup(tccs)
	case *cs_bls12377.SparseR1CS:
		return plonk_bls12377.Setup(tccs)
	case *cs_bw6761.SparseR1CS:
		return plonk_bw6761.Setup(tccs)
	case *cs_bls24315.SparseR1CS:
		return plonk_bls24315.Setup(tccs)
	case *cs_bw6633.SparseR1CS:
		return plonk_bw6633.Setup(tccs)
	case *cs_bls24317.SparseR1CS:
		return plonk_bls24317.Setup(tccs)
	default:
		panic("unrecognized SparseR1CS curve type")
	}

}

// Prove generates PLONK proof from a circuit, associated preprocessed public data, and the witness
// if the force flag is set:
//
//		will executes all the prover computations, even if the witness is invalid
//	 will produce an invalid proof
//		internally, the solution vector to the SparseR1CS will be filled with random values which may impact benchmarking
func Prove(ccs constraint.ConstraintSystem, pk ProvingKey, fullWitness witness.Witness, opts ...backend.ProverOption) (Proof, error) {

	switch tccs := ccs.(type) {
	case *cs_bn254.SparseR1CS:
		return plonk_bn254.Prove(tccs, pk.(*plonk_bn254.ProvingKey), fullWitness, opts...)

	case *cs_bls12381.SparseR1CS:
		return plonk_bls12381.Prove(tccs, pk.(*plonk_bls12381.ProvingKey), fullWitness, opts...)

	case *cs_bls12377.SparseR1CS:
		return plonk_bls12377.Prove(tccs, pk.(*plonk_bls12377.ProvingKey), fullWitness, opts...)

	case *cs_bw6761.SparseR1CS:
		return plonk_bw6761.Prove(tccs, pk.(*plonk_bw6761.ProvingKey), fullWitness, opts...)

	case *cs_bw6633.SparseR1CS:
		return plonk_bw6633.Prove(tccs, pk.(*plonk_bw6633.ProvingKey), fullWitness, opts...)

	case *cs_bls24315.SparseR1CS:
		return plonk_bls24315.Prove(tccs, pk.(*plonk_bls24315.ProvingKey), fullWitness, opts...)

	case *cs_bls24317.SparseR1CS:
		return plonk_bls24317.Prove(tccs, pk.(*plonk_bls24317.ProvingKey), fullWitness, opts...)

	default:
		panic("unrecognized SparseR1CS curve type")
	}
}

// Verify verifies a PLONK proof, from the proof, preprocessed public data, and public witness.
func Verify(proof Proof, vk VerifyingKey, publicWitness witness.Witness, opts ...backend.VerifierOption) error {

	switch _proof := proof.(type) {

	case *plonk_bn254.Proof:
		w, ok := publicWitness.Vector().(fr_bn254.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bn254.Verify(_proof, vk.(*plonk_bn254.VerifyingKey), w, opts...)

	case *plonk_bls12381.Proof:
		w, ok := publicWitness.Vector().(fr_bls12381.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bls12381.Verify(_proof, vk.(*plonk_bls12381.VerifyingKey), w, opts...)

	case *plonk_bls12377.Proof:
		w, ok := publicWitness.Vector().(fr_bls12377.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bls12377.Verify(_proof, vk.(*plonk_bls12377.VerifyingKey), w, opts...)

	case *plonk_bw6761.Proof:
		w, ok := publicWitness.Vector().(fr_bw6761.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bw6761.Verify(_proof, vk.(*plonk_bw6761.VerifyingKey), w, opts...)

	case *plonk_bw6633.Proof:
		w, ok := publicWitness.Vector().(fr_bw6633.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bw6633.Verify(_proof, vk.(*plonk_bw6633.VerifyingKey), w, opts...)

	case *plonk_bls24315.Proof:
		w, ok := publicWitness.Vector().(fr_bls24315.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bls24315.Verify(_proof, vk.(*plonk_bls24315.VerifyingKey), w, opts...)
	case *plonk_bls24317.Proof:
		w, ok := publicWitness.Vector().(fr_bls24317.Vector)
		if !ok {
			return witness.ErrInvalidWitness
		}
		return plonk_bls24317.Verify(_proof, vk.(*plonk_bls24317.VerifyingKey), w, opts...)

	default:
		panic("unrecognized proof type")
	}
}
