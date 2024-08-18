package SplitPipeline

import (
	"Yoimiya/Circuit"
	groth16_bn254 "Yoimiya/backend/groth16/bn254"
	"Yoimiya/constraint"
	cs "Yoimiya/constraint/bn254"
	"Yoimiya/frontend"
	"Yoimiya/frontend/split"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"net"
)

func handleRequest(conn net.Conn, message *string) bool {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return false
	}
	*message = string(buffer[:n])
	fmt.Printf("Received: %s\n", *message)
	conn.Write([]byte("Message received"))
	return true
}

// todo fakeSolve
// 这里简化了下流程，省去了读写文件或solve结果的传输
// prove这边就简单的用同一个solution
// 可以用序列化的方式将每个solve的结果通过tcp发送
func fakeSolve(circuit Circuit.TestCircuit, ccs constraint.ConstraintSystem, witnessID []int, extra []constraint.ExtraValue) SolverSolution {
	assignment := circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)

	witness, err := frontend.GenerateSplitWitnessFromPli(pli, witnessID, extra, ecc.BN254.ScalarField())
	if err != nil {
		panic(err)
	}
	commitmentsInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.SimpleSolve(ccs.(*cs.R1CS), witness)
	if err != nil {
		panic(err)
	}
	return SolverSolution{
		commitmentInfo: commitmentsInfo,
		solution:       solution,
		nbPublic:       nbPublic,
		nbPrivate:      nbPrivate,
	}
}
func fakeSplitSolve(circuit Circuit.TestCircuit, pcs PipelineConstraintSystem) []SolverSolution {
	assignment := circuit.GetAssignment()
	pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	extra := make([]constraint.ExtraValue, 0)
	solverSolutions := make([]SolverSolution, 0)
	for pcs.Next() {
		ccs, witnessID := pcs.Params()
		witness, err := frontend.GenerateSplitWitnessFromPli(pli, witnessID, extra, ecc.BN254.ScalarField())
		if err != nil {
			panic(err)
		}
		commitmentsInfo, solution, nbPublic, nbPrivate, err := groth16_bn254.SimpleSolve(ccs.(*cs.R1CS), witness)
		if err != nil {
			panic(err)
		}
		newExtra := split.GetExtra(ccs)
		extra = append(extra, newExtra...)
		solverSolutions = append(solverSolutions, SolverSolution{
			commitmentInfo: commitmentsInfo,
			solution:       solution,
			nbPublic:       nbPublic,
			nbPrivate:      nbPrivate,
		})
	}
	//ccs, _ := se.pcs.GetParams(input.phase)
	//se.solveLock <- 1
	//startTime := time.Now()
	return solverSolutions
}
