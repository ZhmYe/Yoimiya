package evaluate

import (
	"Yoimiya/Circuit"
	"Yoimiya/Circuit/Circuit4Conv"
	"Yoimiya/Circuit/Circuit4Fib"
	"Yoimiya/Circuit/Circuit4MatrixMultiplication"
	"Yoimiya/Circuit/Circuit4Multiplication"
	"Yoimiya/Circuit/Circuit4VerifyCircuit"
	"Yoimiya/logWriter"
	"Yoimiya/plugin"
	"strconv"
)

// Format 用于format log的名字
func Format(circuitName string, testName string) string {
	return "test_" + testName + "_" + circuitName
}

// GetCircuit 根据电路枚举给出电路
func GetCircuit(option Circuit.CircuitOption) Circuit.TestCircuit {
	switch option {
	case Circuit.Fib:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	case Circuit.FibSquare:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	case Circuit.Mul:
		circuit := Circuit4Multiplication.NewLoopMultiplicationCircuit()
		return &circuit
	case Circuit.Verify:
		circuit := Circuit4VerifyCircuit.NewVerifyCircuit()
		return &circuit
	case Circuit.Matrix:
		circuit := Circuit4MatrixMultiplication.NewInterfaceMatrixMultiplicationCircuit()
		return &circuit
	case Circuit.Conv:
		circuit := Circuit4Conv.NewInterfaceConvolutionalCircuit()
		return &circuit
	default:
		circuit := Circuit4Fib.NewLoopFibonacciCircuit()
		return &circuit
	}
}

func PluginRecordLog(records []plugin.PluginRecord, path string) {
	lw := logWriter.NewLogWriter(path)
	for _, record := range records {
		lw.Writeln("[" + record.Name + " Record]")
		lw.Writeln("\t[Memory Used]: " + strconv.FormatFloat(record.Memory.TotalMemoryUsed, 'f', 8, 64) + " GB")
		lw.Writeln("\t[Prove Memory Used: ]" + strconv.FormatFloat(record.Memory.ProveMemoryUsed, 'f', 8, 64) + " GB")
		for _, pt := range record.Times {
			lw.Writeln("\t[" + pt.Name + "]: " + pt.TimeUsed.String())
			//lw.Writeln("\t[%s]: %v \n", pt.Name, pt.TimeUsed)
		}
	}
	//lw.Writeln("[Record]: ")
	lw.Finish()
}
