package graph

import (
	"fmt"
	"strconv"
)

// Create by ZhmYe 2023/12/18 15:39

type DAG struct {
	list map[int][]int // 邻接表
}

func NewDAG() *DAG {
	d := new(DAG)
	d.list = make(map[int][]int)
	return d
}

func (d *DAG) Size() int {
	return len(d.list)
}
func (d *DAG) Exist(id int) bool {
	_, exist := d.list[id]
	return exist
}
func (d *DAG) SizeOf(id int) int {
	if !d.Exist(id) {
		return 0
	} else {
		return len(d.list[id])
	}
}
func (d *DAG) GetLinks(id int) (links []int) {
	if !d.Exist(id) {
		return links
	} else {
		return d.list[id]
	}
}
func (d *DAG) insert(id int) {
	if d.Exist(id) {
		return
	}
	d.list[id] = make([]int, 0)
}
func (d *DAG) Update(from int, to int) {
	if !d.Exist(from) {
		d.insert(from)
	}
	d.list[from] = append(d.list[from], to)
}
func (d *DAG) Print() {
	for idx, l := range d.list {
		fmt.Print("from: " + strconv.Itoa(idx) + " to: ")
		fmt.Println(l)
	}
}
