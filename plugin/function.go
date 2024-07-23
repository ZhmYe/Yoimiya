package plugin

import (
	cs_bn254 "Yoimiya/constraint/bn254"
	"fmt"
)

func PrintConstraintSystemInfo(cs *cs_bn254.R1CS, name string) {
	fmt.Println("[", name, "]", " Compile Result: ")
	fmt.Println("	NbPublic=", cs.GetNbPublicVariables(), " NbSecret=", cs.GetNbSecretVariables(), " NbInternal=", cs.GetNbInternalVariables())
	fmt.Println("	NbConstraints=", cs.GetNbConstraints())
	fmt.Println("	NbWires=", cs.GetNbPublicVariables()+cs.GetNbSecretVariables()+cs.GetNbInternalVariables())
}
