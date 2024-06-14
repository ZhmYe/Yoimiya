package LRO_Tree

import (
	"reflect"
	"sort"
)

type SortEngine struct {
	sortFunc interface{}
}

func NewSortEngine(sortFunc interface{}) *SortEngine {
	return &SortEngine{sortFunc: sortFunc}
}

func (se *SortEngine) Sort(data interface{}) interface{} {
	dataVal := reflect.ValueOf(data)
	if dataVal.Kind() != reflect.Slice {
		panic("Sort: data must be a slice")
	}

	sortFuncVal := reflect.ValueOf(se.sortFunc)
	elemType := dataVal.Type().Elem()
	lessFuncType := reflect.FuncOf([]reflect.Type{elemType, elemType}, []reflect.Type{reflect.TypeOf(true)}, false)

	if sortFuncVal.Type() != lessFuncType {
		panic("Sort: sort function must be of type func(T, T) bool where T is the element type of the slice")
	}

	sort.Slice(data, func(i, j int) bool {
		return sortFuncVal.Call([]reflect.Value{dataVal.Index(i), dataVal.Index(j)})[0].Bool()
	})

	return data
}
func RoundUpSplit(total int, cut int) int {
	return (total-total%cut)/cut + 1
}
