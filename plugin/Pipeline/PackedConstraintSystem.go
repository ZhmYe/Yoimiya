package Pipeline

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/constraint"
)

type PackedConstraintSystem struct {
	cs      constraint.ConstraintSystem
	pk      groth16.ProvingKey
	vk      groth16.VerifyingKey
	witness []int // 记录该电路哪些input是witness
}
