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

package groth16

import (
	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"

	"Yoimiya/internal/utils"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/pedersen"
	"io"
)

// WriteTo writes binary encoding of the Proof elements to writer
// points are stored in compressed form Ar | Krs | Bs
// use WriteRawTo(...) to encode the proof without point compression
func (proof *Proof) WriteTo(w io.Writer) (n int64, err error) {
	return proof.writeTo(w, false)
}

// WriteRawTo writes binary encoding of the Proof elements to writer
// points are stored in uncompressed form Ar | Krs | Bs
// use WriteTo(...) to encode the proof with point compression
func (proof *Proof) WriteRawTo(w io.Writer) (n int64, err error) {
	return proof.writeTo(w, true)
}

func (proof *Proof) writeTo(w io.Writer, raw bool) (int64, error) {
	var enc *curve.Encoder
	if raw {
		enc = curve.NewEncoder(w, curve.RawEncoding())
	} else {
		enc = curve.NewEncoder(w)
	}

	if err := enc.Encode(&proof.Ar); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&proof.Bs); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&proof.Krs); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(proof.Commitments); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&proof.CommitmentPok); err != nil {
		return enc.BytesWritten(), err
	}

	return enc.BytesWritten(), nil
}

// ReadFrom attempts to decode a Proof from reader
// Proof must be encoded through WriteTo (compressed) or WriteRawTo (uncompressed)
func (proof *Proof) ReadFrom(r io.Reader) (n int64, err error) {

	dec := curve.NewDecoder(r)

	if err := dec.Decode(&proof.Ar); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&proof.Bs); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&proof.Krs); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&proof.Commitments); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&proof.CommitmentPok); err != nil {
		return dec.BytesRead(), err
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of the key elements to writer
// points are compressed
// use WriteRawTo(...) to encode the key without point compression
func (vk *VerifyingKey) WriteTo(w io.Writer) (n int64, err error) {
	if n, err = vk.writeTo(w, false); err != nil {
		return n, err
	}
	var m int64
	m, err = vk.CommitmentKey.WriteTo(w)
	return m + n, err
}

// WriteRawTo writes binary encoding of the key elements to writer
// points are not compressed
// use WriteTo(...) to encode the key with point compression
func (vk *VerifyingKey) WriteRawTo(w io.Writer) (n int64, err error) {
	if n, err = vk.writeTo(w, true); err != nil {
		return n, err
	}
	var m int64
	m, err = vk.CommitmentKey.WriteRawTo(w)
	return m + n, err
}

// writeTo serialization format:
// follows bellman format:
// https://github.com/zkcrypto/bellman/blob/fa9be45588227a8c6ec34957de3f68705f07bd92/src/groth16/mod.rs#L143
// [α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2,uint32(len(Kvk)),[Kvk]1
func (vk *VerifyingKey) writeTo(w io.Writer, raw bool) (int64, error) {
	var enc *curve.Encoder
	if raw {
		enc = curve.NewEncoder(w, curve.RawEncoding())
	} else {
		enc = curve.NewEncoder(w)
	}

	// [α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2
	if err := enc.Encode(&vk.G1.Alpha); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&vk.G1.Beta); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&vk.G2.Beta); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&vk.G2.Gamma); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&vk.G1.Delta); err != nil {
		return enc.BytesWritten(), err
	}
	if err := enc.Encode(&vk.G2.Delta); err != nil {
		return enc.BytesWritten(), err
	}

	// uint32(len(Kvk)),[Kvk]1
	if err := enc.Encode(vk.G1.K); err != nil {
		return enc.BytesWritten(), err
	}

	if vk.PublicAndCommitmentCommitted == nil {
		vk.PublicAndCommitmentCommitted = [][]int{} // only matters in tests
	}
	if err := enc.Encode(utils.IntSliceSliceToUint64SliceSlice(vk.PublicAndCommitmentCommitted)); err != nil {
		return enc.BytesWritten(), err
	}

	return enc.BytesWritten(), nil
}

// ReadFrom attempts to decode a VerifyingKey from reader
// VerifyingKey must be encoded through WriteTo (compressed) or WriteRawTo (uncompressed)
// serialization format:
// https://github.com/zkcrypto/bellman/blob/fa9be45588227a8c6ec34957de3f68705f07bd92/src/groth16/mod.rs#L143
// [α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2,uint32(len(Kvk)),[Kvk]1
func (vk *VerifyingKey) ReadFrom(r io.Reader) (int64, error) {
	n, err := vk.readFrom(r)
	if err != nil {
		return n, err
	}
	var m int64
	m, err = vk.CommitmentKey.ReadFrom(r)
	return m + n, err
}

// UnsafeReadFrom has the same behavior as ReadFrom, except that it will not check that decode points
// are on the curve and in the correct subgroup.
func (vk *VerifyingKey) UnsafeReadFrom(r io.Reader) (int64, error) {
	n, err := vk.readFrom(r, curve.NoSubgroupChecks())
	if err != nil {
		return n, err
	}
	var m int64
	m, err = vk.CommitmentKey.UnsafeReadFrom(r)
	return m + n, err
}

func (vk *VerifyingKey) readFrom(r io.Reader, decOptions ...func(*curve.Decoder)) (int64, error) {
	dec := curve.NewDecoder(r, decOptions...)

	// [α]1,[β]1,[β]2,[γ]2,[δ]1,[δ]2
	if err := dec.Decode(&vk.G1.Alpha); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&vk.G1.Beta); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&vk.G2.Beta); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&vk.G2.Gamma); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&vk.G1.Delta); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&vk.G2.Delta); err != nil {
		return dec.BytesRead(), err
	}

	// uint32(len(Kvk)),[Kvk]1
	if err := dec.Decode(&vk.G1.K); err != nil {
		return dec.BytesRead(), err
	}
	var publicCommitted [][]uint64
	if err := dec.Decode(&publicCommitted); err != nil {
		return dec.BytesRead(), err
	}
	vk.PublicAndCommitmentCommitted = utils.Uint64SliceSliceToIntSliceSlice(publicCommitted)

	// recompute vk.e (e(α, β)) and  -[δ]2, -[γ]2
	if err := vk.Precompute(); err != nil {
		return dec.BytesRead(), err
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of the key elements to writer
// points are compressed
// use WriteRawTo(...) to encode the key without point compression
func (pk *ProvingKey) WriteTo(w io.Writer) (n int64, err error) {
	return pk.writeTo(w, false)
}

// WriteRawTo writes binary encoding of the key elements to writer
// points are not compressed
// use WriteTo(...) to encode the key with point compression
func (pk *ProvingKey) WriteRawTo(w io.Writer) (n int64, err error) {
	return pk.writeTo(w, true)
}

func (pk *ProvingKey) writeTo(w io.Writer, raw bool) (int64, error) {
	n, err := pk.Domain.WriteTo(w)
	if err != nil {
		return n, err
	}

	var enc *curve.Encoder
	if raw {
		enc = curve.NewEncoder(w, curve.RawEncoding())
	} else {
		enc = curve.NewEncoder(w)
	}
	nbWires := uint64(len(pk.InfinityA))

	toEncode := []interface{}{
		&pk.G1.Alpha,
		&pk.G1.Beta,
		&pk.G1.Delta,
		pk.G1.A,
		pk.G1.B,
		pk.G1.Z,
		pk.G1.K,
		&pk.G2.Beta,
		&pk.G2.Delta,
		pk.G2.B,
		nbWires,
		pk.NbInfinityA,
		pk.NbInfinityB,
		pk.InfinityA,
		pk.InfinityB,
		uint32(len(pk.CommitmentKeys)),
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return n + enc.BytesWritten(), err
		}
	}

	for i := range pk.CommitmentKeys {
		var (
			n2  int64
			err error
		)
		if raw {
			n2, err = pk.CommitmentKeys[i].WriteRawTo(w)
		} else {
			n2, err = pk.CommitmentKeys[i].WriteTo(w)
		}

		n += n2
		if err != nil {
			return n, err
		}
	}

	return n + enc.BytesWritten(), nil

}

// ReadFrom attempts to decode a ProvingKey from reader
// ProvingKey must be encoded through WriteTo (compressed) or WriteRawTo (uncompressed)
// note that we don't check that the points are on the curve or in the correct subgroup at this point
func (pk *ProvingKey) ReadFrom(r io.Reader) (int64, error) {
	return pk.readFrom(r)
}

// UnsafeReadFrom behaves like ReadFrom excepts it doesn't check if the decoded points are on the curve
// or in the correct subgroup
func (pk *ProvingKey) UnsafeReadFrom(r io.Reader) (int64, error) {
	return pk.readFrom(r, curve.NoSubgroupChecks())
}

func (pk *ProvingKey) readFrom(r io.Reader, decOptions ...func(*curve.Decoder)) (int64, error) {
	n, err := pk.Domain.ReadFrom(r)
	if err != nil {
		return n, err
	}

	dec := curve.NewDecoder(r, decOptions...)

	var nbWires uint64
	var nbCommitments uint32

	toDecode := []interface{}{
		&pk.G1.Alpha,
		&pk.G1.Beta,
		&pk.G1.Delta,
		&pk.G1.A,
		&pk.G1.B,
		&pk.G1.Z,
		&pk.G1.K,
		&pk.G2.Beta,
		&pk.G2.Delta,
		&pk.G2.B,
		&nbWires,
		&pk.NbInfinityA,
		&pk.NbInfinityB,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return n + dec.BytesRead(), err
		}
	}
	pk.InfinityA = make([]bool, nbWires)
	pk.InfinityB = make([]bool, nbWires)

	if err := dec.Decode(&pk.InfinityA); err != nil {
		return n + dec.BytesRead(), err
	}
	if err := dec.Decode(&pk.InfinityB); err != nil {
		return n + dec.BytesRead(), err
	}
	if err := dec.Decode(&nbCommitments); err != nil {
		return n + dec.BytesRead(), err
	}

	pk.CommitmentKeys = make([]pedersen.ProvingKey, nbCommitments)
	for i := range pk.CommitmentKeys {
		n2, err := pk.CommitmentKeys[i].ReadFrom(r)
		n += n2
		if err != nil {
			return n, err
		}
	}

	return n + dec.BytesRead(), nil
}
