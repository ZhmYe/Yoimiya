package SolveProvePerformance

import (
	"Yoimiya/Circuit"
	"Yoimiya/Config"
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/evaluate"
	"Yoimiya/frontend"
	"Yoimiya/logWriter"
	"Yoimiya/plugin"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Experiment_Solve_Prove_Performance
// SolveProvePerformance Experiment
// 测试Solve和Prove在不同的核数下的时间和CPU利用率

// Experiment_Solve_Prove_Time_Performance 不同核数时间测试
func Experiment_Solve_Prove_Time_Performance(option Circuit.CircuitOption, log bool) {
	Config.Config.CancelSplit()
	circuit := evaluate.GetCircuit(option)
	var lw *logWriter.LogWriter
	if log {
		lw = logWriter.NewLogWriter("SolveProvePerformance" + "/" + "record_log_" + evaluate.Format(circuit.Name(), "time_performance"))
		lw.Writeln("[Record]: ")
	} else {
		fmt.Println("[Record]")
	}
	NumCPUList := []int{64, 32, 16, 8, 4, 2, 1}
	ccs, _ := circuit.Compile()
	assignment := circuit.GetAssignment()
	//pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	fullWitness, _ := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())
	//publicWitness, err := fullWitness.Public()
	//if err != nil {
	//	panic(err)
	//}
	pk, _ := frontend.SetUpSplit(ccs)
	TimePerformance := func(ccs constraint.ConstraintSystem, pk groth16.ProvingKey, witness witness.Witness,
		NumCPU int, log bool) {
		runtime.GOMAXPROCS(NumCPU)
		prover := plugin.NewProver(pk)
		totalSolveTime := time.Duration(0)
		totalProveTime := time.Duration(0)
		for i := 0; i < 10; i++ {
			solveStartTime := time.Now()
			commitmentsInfo, solution, nbPublic, nbPrivate := prover.Solve(ccs.(*cs_bn254.R1CS), witness)
			solveTime := time.Since(solveStartTime)
			proveStartTime := time.Now()
			_, err := prover.Prove(*solution, commitmentsInfo, nbPublic, nbPrivate)
			if err != nil {
				return
			}
			proveTime := time.Since(proveStartTime)
			totalSolveTime += solveTime
			totalProveTime += proveTime
		}
		totalProveTime /= 10
		totalSolveTime /= 10
		if log {
			lw.Writeln("	[NumCPU]: " + strconv.Itoa(NumCPU))
			lw.Writeln("		[Solve Time]: " + totalSolveTime.String())
			lw.Writeln("		[Prove Time]: " + totalProveTime.String())
		} else {
			fmt.Println("	[NumCPU]: " + strconv.Itoa(NumCPU))
			fmt.Println("		[Solve Time]: " + totalSolveTime.String())
			fmt.Println("		[Prove Time]: " + totalProveTime.String())
		}
	}
	for _, NumCPU := range NumCPUList {
		TimePerformance(ccs, pk, fullWitness, NumCPU*2, log)
	}
	lw.Finish()
}
func Experiment_Solve_Prove_CPU_Performance(option Circuit.CircuitOption, log bool) {
	Config.Config.CancelSplit()
	circuit := evaluate.GetCircuit(option)
	var lw *logWriter.LogWriter
	if log {
		lw = logWriter.NewLogWriter("SolveProvePerformance" + "/" + "record_log_" + evaluate.Format(circuit.Name(), "cpu_performance"))
		lw.Writeln("[Record]: ")
	} else {
		fmt.Println("[Record]")
	}
	NumCPUList := []int{64, 32, 16, 8, 4, 2, 1}
	ccs, _ := circuit.Compile()
	assignment := circuit.GetAssignment()
	//pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	fullWitness, _ := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())
	//publicWitness, err := fullWitness.Public()
	//if err != nil {
	//	panic(err)
	//}
	pk, _ := frontend.SetUpSplit(ccs)
	CPUPerformance := func(ccs constraint.ConstraintSystem, pk groth16.ProvingKey, witness witness.Witness,
		NumCPU int, log bool) {
		//if NumCPU > runtime.NumCPU() {
		//	NumCPU = runtime.NumCPU()
		//}
		NumCPU = runtime.GOMAXPROCS(NumCPU)
		prover := plugin.NewProver(pk)
		totalSolveCPUUsage := float64(0)
		totalProveCPUUsage := float64(0)
		for i := 0; i < 1; i++ {
			monitor := evaluate.NewMonitor(true, false)
			monitor.Start()
			prover.Solve(ccs.(*cs_bn254.R1CS), witness)
			record := monitor.Finish()
			totalSolveCPUUsage += record.CPUUsage()
		}
		// 用于下一次prove
		commitmentsInfo, solution, nbPublic, nbPrivate := prover.Solve(ccs.(*cs_bn254.R1CS), witness)
		for i := 0; i < 1; i++ {
			monitor := evaluate.NewMonitor(true, false)
			monitor.Start()
			_, err := prover.Prove(*solution, commitmentsInfo, nbPublic, nbPrivate)
			if err != nil {
				return
			}
			record := monitor.Finish()
			totalProveCPUUsage += record.CPUUsage()
		}
		totalSolveCPUUsage /= 1
		totalProveCPUUsage /= 1
		if log {
			lw.Writeln("	[NumCPU]: " + strconv.Itoa(NumCPU))
			lw.Writeln("		[Solve CPU Usage]: " + strconv.FormatFloat(totalSolveCPUUsage*float64(runtime.NumCPU())/float64(NumCPU), 'f', 2, 64))
			lw.Writeln("		[Prove CPU Usage]: " + strconv.FormatFloat(totalProveCPUUsage*float64(runtime.NumCPU())/float64(NumCPU), 'f', 2, 64))
		} else {
			fmt.Println("	[NumCPU]: " + strconv.Itoa(NumCPU))
			fmt.Println("		[Solve CPU Usage]:", totalSolveCPUUsage)
			fmt.Println("		[Prove CPU Usage]:", totalProveCPUUsage)
		}

	}
	for _, NumCPU := range NumCPUList {
		CPUPerformance(ccs, pk, fullWitness, NumCPU*2, log)
	}
	lw.Finish()
}

func Experiment_Solve_Performance(option Circuit.CircuitOption, log bool) {
	Config.Config.CancelSplit()
	circuit := evaluate.GetCircuit(option)
	//fmt.Println("[Record]")
	//NumCPUList := []int{64, 32, 16, 8, 4, 2, 1}
	ccs, _ := circuit.Compile()
	assignment := circuit.GetAssignment()
	//pli := frontend.GetPackedLeafInfoFromAssignment(assignment)
	fullWitness, _ := frontend.GenerateWitness(assignment, make([]constraint.ExtraValue, 0), ecc.BN254.ScalarField())
	pk, _ := frontend.SetUpSplit(ccs)
	TimePerformance := func(ccs constraint.ConstraintSystem, pk groth16.ProvingKey, witness witness.Witness,
		NumCPU int, log bool) {
		runtime.GOMAXPROCS(NumCPU)
		prover := plugin.NewProver(pk)
		for {
			var wg sync.WaitGroup
			wg.Add(8)
			runtime.LockOSThread()
			for i := 0; i < 8; i++ {
				go func() {
					solveStartTime := time.Now()
					prover.Solve(ccs.(*cs_bn254.R1CS), witness)
					solveTime := time.Since(solveStartTime)
					fmt.Println(solveTime)
					wg.Done()
				}()
			}
			wg.Wait()
		}

	}
	//for _, NumCPU := range NumCPUList {
	TimePerformance(ccs, pk, fullWitness, runtime.NumCPU()/2, log)
	//}
	//lw.Finish()
}
