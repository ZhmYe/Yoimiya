package frontend

import (
	"S-gnark/backend/groth16"
	"S-gnark/backend/witness"
	"S-gnark/constraint"
	cs_bn254 "S-gnark/constraint/bn254"
	"S-gnark/frontend/schema"
	"errors"
	"math/big"
	"reflect"
	"strconv"
)

// SetNbLeaf 设置nbPublic、nbSecret
// todo 这里还需要加上extra
func SetNbLeaf(assignment Circuit, cs *cs_bn254.R1CS, extra []constraint.ExtraValue) error {
	(*cs).AddPublicVariable("1")
	idx := (*cs).GetNbWires() - 1
	(*cs).SetBias(uint32(idx), idx)
	variableAdder := func() func(f schema.LeafInfo, tInput reflect.Value) error {
		return func(f schema.LeafInfo, tInput reflect.Value) error {
			if tInput.CanSet() {
				if f.Visibility == schema.Unset {
					return errors.New("can't set val " + f.FullName() + " visibility is unset")
				}
				if f.Visibility == schema.Public {
					(*cs).AddPublicVariable(f.FullName())
				} else if f.Visibility == schema.Secret {
					(*cs).AddSecretVariable(f.FullName())
				}
			}
			// 这里后续witness的排序是按照先public然后private的顺序来的，下面有extra作为public，因此private的bias应该要靠后
			if f.Visibility == schema.Public {
				// 如果是public，那么wireID和bias是一样的
				idx := (*cs).GetNbWires() - 1
				(*cs).SetBias(uint32(idx), idx)
			} else if f.Visibility == schema.Secret {
				// 如果是secret， 那么bias上还需要加上extra的数量
				idx := (*cs).GetNbWires() - 1
				(*cs).SetBias(uint32(idx), idx+len(extra))
			}
			return nil
		}
		//return errors.New("can't set val ")
	}
	_, err := schema.Walk(assignment, tVariable, variableAdder())
	if err != nil {
		return err
	}
	// 这里设置了extra的偏移
	//fmt.Println("len ForwardOutput", len(cs.GetForwardOutputs()))
	for _, e := range extra {
		//if e.IsUsed() {
		//	continue
		//}
		(*cs).AddPublicVariable("ForwardOutput_" + strconv.Itoa(e.GetWireID())) // 这里设置为public，因为上半的输出应该是公开的，另外也简化了public witness的生成
		// 这里应该要看的是public的数量
		idx := (*cs).GetNbPublicVariables() - 1
		(*cs).SetBias(uint32(e.GetWireID()), idx)
	}
	(*cs).SetExtraNumber(extra)
	return nil
}

// todo 如何得到extra，即如何得到MIDDLE的值
// GenerateWitness 为split后的电路生成witness,extra表示Middle

func GenerateWitness(assignment Circuit, extra []constraint.ExtraValue, field *big.Int, opts ...WitnessOption) (witness.Witness, error) {
	opt, err := options(opts...)
	if err != nil {
		return nil, err
	}
	// todo 这里split出的电路已经无法支持schema.Walk，但我们已知输入就是MIDDLE（可能还有之前原电路的input?）
	// 将所有的MIDDLE认为是Public Input，直接统计得到NbPublic和NbPrivate
	// 我们可以保留原电路的Circuit，这应该不占太多内存(?)，这还没有解析到整个cs，所以只有Input
	// 这里直接传入assignment,通过原有代码得到原本的public Input和private Input
	// todo 是否需要得到目前需要的input，而不是所有Input
	// count the leaves
	s, err := schema.Walk(assignment, tVariable, nil)
	if err != nil {
		return nil, err
	}
	if opt.publicOnly {
		s.Secret = 0
	}

	// allocate the witness
	w, err := witness.New(field)
	if err != nil {
		return nil, err
	}
	extraNumber := len(extra)
	//for _, e := range extra {
	//if !e.IsUsed() {
	//extraNumber++
	//}
	//}
	// write the public | secret values in a chan
	chValues := make(chan any)
	go func() {
		defer close(chValues)
		schema.Walk(assignment, tVariable, func(leaf schema.LeafInfo, tValue reflect.Value) error {
			if leaf.Visibility == schema.Public {
				chValues <- tValue.Interface()
			}
			return nil
		})
		// todo 这里不确定是否这样写
		// 传入MIDDLE的值作为Input
		// 这里因为extra作为public，所以按顺序应该在这里
		for _, e := range extra {
			//if e.IsUsed() {
			//	continue
			//}
			chValues <- e.GetValue()
		}
		if !opt.publicOnly {
			schema.Walk(assignment, tVariable, func(leaf schema.LeafInfo, tValue reflect.Value) error {
				if leaf.Visibility == schema.Secret {
					chValues <- tValue.Interface()
				}
				return nil
			})
		}
	}()
	if err := w.Fill(s.Public+extraNumber, s.Secret, chValues); err != nil {
		return nil, err
	}

	return w, nil
}

// SetUpSplit 给定电路 进行SetUp操作并给出ProveingKey和VerifyingKey
func SetUpSplit(cs constraint.ConstraintSystem) (groth16.ProvingKey, groth16.VerifyingKey) {
	//startTime := time.Now()
	outerPK, outerVK, err := groth16.Setup(cs)
	if err != nil {
		panic(err)
	}
	//fmt.Println("SetUp Time:", time.Since(startTime))
	//full, err := NewWitness(outerAssignment, ecc.BN254.ScalarField())
	//public, err := NewWitness(outerAssignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	return outerPK, outerVK
}
