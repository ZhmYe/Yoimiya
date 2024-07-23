package frontend

import (
	"Yoimiya/backend/groth16"
	"Yoimiya/backend/witness"
	"Yoimiya/constraint"
	cs_bn254 "Yoimiya/constraint/bn254"
	"Yoimiya/frontend/schema"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"
)

// PackedLeafInfo 这里直接统一得到assignment得到的public和private信息

type PackedValue struct {
	name       string
	value      reflect.Value
	visibility bool
}

func (v *PackedValue) IsPublic() bool {
	return v.visibility
}

type PackedLeafInfo struct {
	public  []PackedValue
	private []PackedValue
}

func (i *PackedLeafInfo) AddPublic(name string, value reflect.Value) {
	//i.value = append(i.value, PackedValue{
	//	name:       name,
	//	visibility: true,
	//})
	//i.nbPublic++
	i.public = append(i.public, PackedValue{name: name, value: value, visibility: true})
}
func (i *PackedLeafInfo) AddSecret(name string, value reflect.Value) {
	//i.value = append(i.value, PackedValue{
	//	name:       name,
	//	visibility: false,
	//})
	i.private = append(i.private, PackedValue{name: name, value: value, visibility: false})
}
func (i *PackedLeafInfo) NbPublic() int {
	//return i.nbPublic

	return len(i.public)
}

func (i *PackedLeafInfo) NbSecret() int {
	return len(i.private)
}
func (i *PackedLeafInfo) GetPublicVariables() []PackedValue {
	return i.public
}
func (i *PackedLeafInfo) GetSecretVariables() []PackedValue {
	return i.private
}
func (i *PackedLeafInfo) GetVariablesByWireID(wireID int) PackedValue {
	if wireID < i.NbPublic() {
		return i.public[wireID]
	} else {
		return i.private[wireID-i.NbPublic()]
	}
}

//func (i *PackedLeafInfo) Copy() PackedLeafInfo {
//	p := make([]string, len(i.PUBLIC))
//	copy(p, i.PUBLIC)
//	s := make([]string, len(i.SECRET))
//	copy(s, i.SECRET)
//	return PackedLeafInfo{
//		PUBLIC: p,
//		SECRET: s,
//	}
//}

func NewPackedLeafInfo() PackedLeafInfo {
	return PackedLeafInfo{
		//value:    make([]PackedValue, 0),
		//nbPublic: 0,
		public:  []PackedValue{{name: "1"}},
		private: make([]PackedValue, 0),
	}
}

// GetPackedLeafInfoFromAssignment 这里从assignment中获取原电路的输入，包括赋值，存在packedLeafInfo中
// 用于替换GetNbLeaf
func GetPackedLeafInfoFromAssignment(assignment Circuit) PackedLeafInfo {
	pli := NewPackedLeafInfo()
	_, err := schema.Walk(assignment, tVariable, func(leaf schema.LeafInfo, tValue reflect.Value) error {
		if leaf.Visibility == schema.Public {
			pli.AddPublic(leaf.FullName(), tValue)
		} else if leaf.Visibility == schema.Secret {
			//chValues <- tValue.Interface()
			pli.AddSecret(leaf.FullName(), tValue)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return pli
}

//func GetNbLeaf(assignment Circuit) PackedLeafInfo {
//	pli := NewPackedLeafInfo()
//	//pli.AddPublic("1")
//	variableAdder := func() func(f schema.LeafInfo, tInput reflect.Value) error {
//		return func(f schema.LeafInfo, tInput reflect.Value) error {
//			if tInput.CanSet() {
//				if f.Visibility == schema.Unset {
//					return errors.New("can't set val " + f.FullName() + " visibility is unset")
//				}
//				if f.Visibility == schema.Public {
//					pli.AddPublic(f.FullName())
//				} else if f.Visibility == schema.Secret {
//					pli.AddSecret(f.FullName())
//				}
//			}
//			return nil
//		}
//		//return errors.New("can't set val ")
//	}
//	_, err := schema.Walk(assignment, tVariable, variableAdder())
//	if err != nil {
//		panic(err)
//	}
//	return pli
//}

// SetInputVariable ibr中已经存了当前split所涉及的input wires，因此可以直接设置
func SetInputVariable(pli PackedLeafInfo, ibr constraint.IBR, cs *cs_bn254.R1CS, extra []constraint.ExtraValue) error {
	// 这里的ibr里包含的witness是有序的wireID，但是不知道具体是public还是private，因此需要pli辅助
	input := ibr.GetWitness()
	getInputInfo := func(pli PackedLeafInfo, wireID int) (bool, string) {
		nbPublic := pli.NbPublic()
		if wireID >= nbPublic {
			return false, pli.private[wireID-nbPublic].name
		}
		return true, pli.public[wireID].name
	}
	(*cs).AddPublicVariable("1")
	(*cs).SetBias(0, 0)
	for _, v := range input {
		if isPublic, name := getInputInfo(pli, v); isPublic {
			(*cs).AddPublicVariable(name)
			idx := (*cs).GetNbWires() - 1
			(*cs).SetBias(uint32(v), idx)
		} else {
			(*cs).AddSecretVariable(name)
			idx := (*cs).GetNbWires() - 1
			(*cs).SetBias(uint32(v), idx+len(extra))
		}
	}
	for _, e := range extra {
		//if e.IsUsed() {
		//	continue
		//}
		(*cs).AddPublicVariable("ForwardOutput_" + strconv.Itoa(e.GetWireID())) // 这里设置为public，因为上半的输出应该是公开的，另外也简化了public witness的生成
		// 这里应该要看的是public的数量
		idx := (*cs).GetNbPublicVariables() - 1
		(*cs).SetBias(uint32(e.GetWireID()), idx)
	}
	//(*cs).SetExtraNumber(extra)
	return nil
}

// SetNbLeaf 设置nbPublic、nbSecret
// todo 这里还需要加上extra
func SetNbLeaf(pli PackedLeafInfo, cs *cs_bn254.R1CS, extra []constraint.ExtraValue) error {
	//for _, v := range pli.value {
	//	if v.IsPublic() {
	//		(*cs).AddPublicVariable(v.name)
	//		idx := (*cs).GetNbWires() - 1
	//		(*cs).SetBias(uint32(idx), idx)
	//	} else {
	//		(*cs).AddSecretVariable(v.name)
	//		idx := (*cs).GetNbWires() - 1
	//		(*cs).SetBias(uint32(idx), idx+len(extra))
	//	}
	//}
	for _, pubV := range pli.public {
		(*cs).AddPublicVariable(pubV.name)
		idx := (*cs).GetNbWires() - 1
		(*cs).SetBias(uint32(idx), idx)
	}
	for _, priV := range pli.private {
		(*cs).AddSecretVariable(priV.name)
		idx := (*cs).GetNbWires() - 1
		(*cs).SetBias(uint32(idx), idx+len(extra))
	}
	for _, e := range extra {
		//if e.IsUsed() {
		//	continue
		//}
		(*cs).AddPublicVariable("ForwardOutput_" + strconv.Itoa(e.GetWireID())) // 这里设置为public，因为上半的输出应该是公开的，另外也简化了public witness的生成
		// 这里应该要看的是public的数量
		idx := (*cs).GetNbPublicVariables() - 1
		(*cs).SetBias(uint32(e.GetWireID()), idx)
	}
	//(*cs).SetExtraNumber(extra)
	return nil
}

// GenerateSplitWitnessFromPli 在pli中已经得到了input的value, 结合每个Split得到的splitWitness即可得到input部分需要哪些
func GenerateSplitWitnessFromPli(pli PackedLeafInfo, splitInfo []int, extra []constraint.ExtraValue, field *big.Int) (witness.Witness, error) {
	nbPublic := 0
	nbSecret := 0
	// todo 这里可以优化
	// 这里把1去掉
	//fmt.Println(splitInfo)
	for i, wireID := range splitInfo {
		packedValue := pli.GetVariablesByWireID(wireID)
		//fmt.Println(packedValue.name)
		if !packedValue.IsPublic() {
			nbPublic = i
			break
		}
		//chValues <- packedValue.value.Interface()
		//nbPublic++
	}
	nbSecret = len(splitInfo) - nbPublic
	//fmt.Println(nbPublic, nbSecret)
	// todo 这里split出的电路已经无法支持schema.Walk，但我们已知输入就是MIDDLE（可能还有之前原电路的input?）
	// 将所有的MIDDLE认为是Public Input，直接统计得到NbPublic和NbPrivate
	// 我们可以保留原电路的Circuit，这应该不占太多内存(?)，这还没有解析到整个cs，所以只有Input
	// 这里直接传入assignment,通过原有代码得到原本的public Input和private Input
	// todo 是否需要得到目前需要的input，而不是所有Input
	// count the leaves
	//s, err := schema.Walk(assignment, tVariable, nil)
	//if err != nil {
	//	return nil, err
	//}
	//if opt.publicOnly {
	//	s.Secret = 0
	//}

	// allocate the witness
	w, err := witness.New(field)
	if err != nil {
		panic(err)
	}
	//publicWitness, err := witness.New(field)
	//if err != nil {
	//	panic(err)
	//}
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
		privateIndex := 0
		for i, wireID := range splitInfo {
			packedValue := pli.GetVariablesByWireID(wireID)
			if !packedValue.IsPublic() {
				privateIndex = i
				break
			}
			chValues <- packedValue.value.Interface()
			//nbPublic++
		}
		for _, e := range extra {
			//if e.IsUsed() {
			//	continue
			//}
			chValues <- e.GetValue()
		}
		for i := privateIndex; i < len(splitInfo); i++ {
			wireID := splitInfo[i]
			packedValue := pli.GetVariablesByWireID(wireID)
			chValues <- packedValue.value.Interface()
			//nbSecret++
		}

	}()
	//go func() {
	//	defer close(chValues)
	//	//schema.Walk(assignment, tVariable, func(leaf schema.LeafInfo, tValue reflect.Value) error {
	//	//	if leaf.Visibility == schema.Public {
	//	//		chValues <- tValue.Interface()
	//	//	}
	//	//	return nil
	//	//})
	//	for _, v := range pli.GetPublicVariables() {
	//		chValues <- v.value.Interface()
	//	}
	//	// todo 这里不确定是否这样写
	//	// 传入MIDDLE的值作为Input
	//	// 这里因为extra作为public，所以按顺序应该在这里
	//	for _, e := range extra {
	//		//if e.IsUsed() {
	//		//	continue
	//		//}
	//		chValues <- e.GetValue()
	//	}
	//	for _, v := range pli.GetSecretVariables() {
	//		chValues <- v.value.Interface()
	//	}
	//	//if !opt.publicOnly {
	//	//	schema.Walk(assignment, tVariable, func(leaf schema.LeafInfo, tValue reflect.Value) error {
	//	//		if leaf.Visibility == schema.Secret {
	//	//			chValues <- tValue.Interface()
	//	//		}
	//	//		return nil
	//	//	})
	//	//}
	//}()
	if err := w.Fill(nbPublic+extraNumber, nbSecret, chValues); err != nil {
		panic(err)
	}

	return w, nil
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
	startTime := time.Now()
	outerPK, outerVK, err := groth16.Setup(cs)
	fmt.Println("Setup Time: ", time.Since(startTime))
	if err != nil {
		panic(err)
	}
	//full, err := NewWitness(outerAssignment, ecc.BN254.ScalarField())
	//public, err := NewWitness(outerAssignment, ecc.BN254.ScalarField(), frontend.PublicOnly())
	return outerPK, outerVK
}
