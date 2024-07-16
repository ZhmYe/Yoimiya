package Circuir4Mimc

import (
	"Yoimiya/frontend"
	"Yoimiya/std/hash/mimc"
)

type MimcCircuit struct {
	// struct tag on a variable is optional
	// default uses variable name and secret visibility.
	PreImage frontend.Variable
	Hash     frontend.Variable `gnark:",public"`
}

// Define declares the circuit's constraints
// Hash = mimc(PreImage)
func (circuit *MimcCircuit) Define(api frontend.API) error {
	// hash function
	mimc, _ := mimc.NewMiMC(api)
	//sha256, _ := sha3.New384(api)
	//sha256.Write(circuit.PreImage)
	// specify constraints
	// mimc(preImage) == hash
	mimc.Write(circuit.PreImage)
	api.AssertIsEqual(circuit.Hash, mimc.Sum())

	return nil
}
