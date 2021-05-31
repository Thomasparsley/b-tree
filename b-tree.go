package main

import (
	"errors"
	"reflect"
)

type node struct {
	parent    *node   // rodič
	T         int     //
	length    int     //
	keys      []int   // klíč
	childrens []*node // potomci
	isLeaf    bool    // jsem listem
}

func NewBtreeNode(t int, isLeaf bool) *node {
	return &node{
		T:         t,
		length:    0,
		keys:      make([]int, 2*t-1),
		childrens: make([]*node, 2*t),
		isLeaf:    isLeaf,
	}
}

func (this node) GetFirstKey() int {
	if this.length == 0 {
		return -1
	} else {
		return this.keys[0]
	}
}

func (this node) GetFirstKeyIndex() int {

	if this.length == 0 {
		return -1
	} else {
		return 0
	}
}

func (this node) GetLastKey() int {
	if this.length == 0 {
		return -1
	} else {
		return this.keys[this.length-1]
	}
}

func (this node) GetLastKeyIndex() int {

	if this.length == 0 {
		return -1
	} else {
		return this.length - 1
	}
}

func (this node) GetLastChildrenIndex() int {
	for i := len(this.childrens) - 1; i >= 0; i-- {
		if this.childrens[i] != nil {
			return i
		}
	}

	return -1
}

func (this node) GetLastChildren() *node {
	for i := len(this.childrens) - 1; i >= 0; i-- {
		if this.childrens[i] != nil {
			return this.childrens[i]
		}
	}

	return nil
}

func (this node) FindMaxKey() (int, int) {
	if this.isLeaf {
		return this.GetLastKey(), this.GetFirstKeyIndex()
	} else {
		return this.GetLastChildren().FindMaxKey()
	}
}

func (this node) InsertInOrder(key int) {
	if this.length == 0 {
		this.keys[0] = key
	} else {
		i := this.length - 1

		for i >= 0 && key < this.keys[i] {
			this.keys[i+1] = this.keys[i]
			i--
		}

		this.keys[i+1] = key
	}

	this.length++
}

func (this node) InsertNonFull(key int) {
	if this.isLeaf {
		this.InsertInOrder(key)
	} else {
		i := this.length - 1

		for i >= 0 && this.keys[i] > key {
			i--
		}

		i++

		if this.childrens[i].length == 2*this.T-1 {
			this.SplitChild(i)

			if key > this.keys[i] {
				i++
			}
		}

		this.childrens[i].InsertNonFull(key)
	}
}

func (this node) SplitChild(index int) {
	y := this.childrens[index]
	z := NewBtreeNode(this.T, y.isLeaf)

	z.parent = y.parent
	z.length = this.T - 1
	y.length = this.T - 1

	// 1.
	// Zkopírování pravé částí klíčů do nového uzlu
	for j := 0; j < this.T; j++ {
		z.keys[j] = y.keys[j+this.T]
		y.keys[j+this.T] = 0
	}

	if !y.isLeaf {
		// Zkopírování pravé částí potomků do nového uzlu
		for j := 0; j < this.T; j++ {
			z.childrens[j] = y.childrens[j+this.T]
			y.childrens[j+this.T] = nil

			// Potomci v novém uzlu budou mít správného rodiče
			z.childrens[j].parent = z
		}
	}

	// 2.
	// Posuneme klíče v aktuálním uzlu o 1 doprava

	for j := this.length - 1; j >= index; j-- {
		this.keys[j+1] = this.keys[j]
	}

	// Posuneme potomky v aktuálním uzlu o 1 doprava
	for j := this.length; j >= index+1; j-- {
		this.childrens[j+1] = this.childrens[j]
	}

	// 3.
	// Do akzuálního uzlo zapojíme nový uzel
	this.childrens[index+1] = z
	this.keys[index] = y.keys[this.T-1]
	y.keys[this.T-1] = 0
	this.length++
}

func (this node) Search(key int) (*node, int) {
	i := 0

	for i < this.length && key > this.keys[i] {
		i++
	}

	if i < this.length && key == this.keys[i] {
		return &this, i
	}

	if this.isLeaf {
		return nil, -1
	}

	return this.childrens[i].Search(key)
}

func (this node) InsertThisNode(mergedNode *node) {
	length := mergedNode.length
	for i := 0; i < this.length; i++ {
		mergedNode.keys[length+i] = this.keys[i]
		this.keys[i] = 0
		mergedNode.length++
	}

	for i := 0; i < this.GetLastChildrenIndex(); i++ {
		mergedNode.childrens[i] = this.childrens[i]
		this.childrens[i] = nil
	}
}

func (this node) InsertSiblig(mergedNode *node, leftSibling bool, thisNodeIndex int) {
	var siblingNode *node
	if !leftSibling {
		siblingNode = this.parent.childrens[thisNodeIndex+1]
	} else {
		siblingNode = this.parent.childrens[thisNodeIndex-1]
	}

	length := mergedNode.length
	for i := 0; i < siblingNode.length; i++ {
		mergedNode.keys[length+i] = siblingNode.keys[i]
		siblingNode.keys[i] = 0
		mergedNode.length++
	}

	for i := 0; i < siblingNode.GetLastChildrenIndex(); i++ {
		mergedNode.childrens[length+i] = siblingNode.childrens[i]
		siblingNode.childrens[i] = nil
	}
}

func (this node) InsertParentKey(mergedNode *node, leftSibling bool, thisNodeIndex int) {
	var index int

	if !leftSibling {
		index = thisNodeIndex
	} else {
		index = thisNodeIndex - 1
	}

	mergedNode.keys[mergedNode.length] = this.parent.keys[index]

	for i := index; i < 2*this.T-1; i++ {
		if this.parent.keys[i] == 0 {
			break
		}

		this.parent.keys[i] = this.parent.keys[i+1]
	}

	mergedNode.length++
}

func (this node) MoveParentChildrens(startIndex int) {
	for i := startIndex; i < 2*this.T-1; i++ {
		this.parent.childrens[i] = this.parent.childrens[i+1]
	}
}

func (this node) MargeTwoNodes(leftSibling bool, thisNodeIndex int) {
	mergedNode := NewBtreeNode(this.T, this.isLeaf)
	mergedNode.parent = this.parent

	if !leftSibling {
		this.InsertThisNode(mergedNode)
		this.InsertParentKey(mergedNode, leftSibling, thisNodeIndex)
		this.InsertSiblig(mergedNode, leftSibling, thisNodeIndex)

		this.parent.childrens[thisNodeIndex] = mergedNode
		this.MoveParentChildrens(thisNodeIndex + 1)
	} else {
		this.InsertSiblig(mergedNode, leftSibling, thisNodeIndex)
		this.InsertParentKey(mergedNode, leftSibling, thisNodeIndex)
		this.InsertThisNode(mergedNode)

		this.parent.childrens[thisNodeIndex-1] = mergedNode
		this.MoveParentChildrens(thisNodeIndex)
	}
}

func (this node) Delete(index int) {
	this.DeletePhaseOne(index)
	this.DeletePhaseTwo()
}

func (this node) DeletePhaseOne(index int) int {
	if this.isLeaf {
		return this.DeleteFromLeaf(index)
	} else {
		return this.DeleteFromNoLeaf(index)
	}
}

func (this node) DeleteFromLeaf(index int) int {
	temp := this.keys[index]

	for i := index; i < this.length-1; i++ {
		this.keys[i] = this.keys[i+1]
	}

	this.length--

	for i := this.length; i < 2*this.T-1; i++ {
		if this.keys[i] == 0 {
			break
		}

		this.keys[i] = 0
	}

	return temp
}

func (this node) DeleteFromNoLeaf(index int) int {
	temp := this.keys[index]

	key, _ := this.childrens[index].FindMaxKey()
	node, i := this.Search(key)
	this.keys[index] = key

	node.Delete(i)

	return temp
}

func (this node) MoveKeyOverParent(leftSibling bool, thisNodeIndex int) {
	var siblingIndex int
	var siblingKeyIndex int
	var keyIndex int

	if leftSibling {
		siblingIndex = thisNodeIndex - 1
		siblingKeyIndex = this.parent.childrens[siblingIndex].GetLastChildrenIndex()
		keyIndex = thisNodeIndex - 1
	} else {
		siblingIndex = thisNodeIndex + 1
		siblingKeyIndex = this.parent.childrens[siblingIndex].GetFirstKeyIndex()
		keyIndex = thisNodeIndex
	}

	this.InsertInOrder(this.parent.keys[keyIndex])

	for i := this.GetLastChildrenIndex(); i < 2*this.T-1; i++ {
		/* this.childrens[i] = this.childrens[i] */
	}

	// Přelití posledního potomka sourozence
	this.childrens[0] = this.parent.childrens[siblingIndex].GetLastChildren()
	this.parent.childrens[siblingIndex].childrens[this.GetLastChildrenIndex()] = nil

	this.parent.keys[keyIndex] = this.parent.childrens[siblingIndex].DeletePhaseOne(siblingKeyIndex)
}

func (this node) IndexOf() int {
	for i := 0; i < len(this.childrens); i++ {
		if reflect.DeepEqual(this.childrens[i], this) {
			return i
		}
	}

	return -1
}

func (this node) DeletePhaseTwo() {
	// 1.
	if this.parent == nil || this.length >= this.T-1 {
		return
	}

	thisNodeIndex := this.IndexOf()

	// 2.
	// Levý sourozenec
	if thisNodeIndex > 0 && this.parent.childrens[thisNodeIndex-1].length > this.T-1 {
		this.MoveKeyOverParent(true, thisNodeIndex)
		return

		// Pravý sourozenec
	} else if thisNodeIndex < 2*this.T && this.parent.childrens[thisNodeIndex+1] != nil && this.parent.childrens[thisNodeIndex+1].length > this.T-1 {
		this.MoveKeyOverParent(false, thisNodeIndex)
		return
	}

	// 3.
	// Pravý sourozenec
	if thisNodeIndex >= 0 && thisNodeIndex < 2*this.T && this.parent.childrens[thisNodeIndex+1] != nil && this.parent.childrens[thisNodeIndex+1].length >= this.T-1 {
		this.MargeTwoNodes(false, thisNodeIndex)
		this.parent.DeletePhaseTwo()

		// Levý sourozenec
	} else if thisNodeIndex > 0 && thisNodeIndex <= 2*this.T && this.parent.childrens[thisNodeIndex-1].length >= this.T-1 {
		this.MargeTwoNodes(true, thisNodeIndex)
		this.parent.DeletePhaseTwo()
	}
}

type tree struct {
	T    int   //
	root *node // kořenový uzel
}

func NewBtree(t int) tree {
	if t < 2 {
		panic(errors.New("t cannot be less than 2"))
	}

	return tree{
		T:    t,
		root: NewBtreeNode(t, true),
	}
}

func (this tree) Insert(key int) {
	root := this.root

	if root.length == 0 {
		root.InsertInOrder(key)
	} else if root.length == 2*this.T-1 {
		var newNode = NewBtreeNode(this.T, false)
		newNode.childrens[0] = root
		this.root = newNode
		root.parent = newNode

		newNode.SplitChild(0)
		newNode.InsertNonFull(key)
	} else {
		root.InsertNonFull(key)
	}
}

func (this tree) Search(key int) (*node, int) {
	return this.root.Search(key)
}

func (this tree) FindMaxKey() (int, int) {
	return this.root.FindMaxKey()
}

func (this tree) Delete(key int) int {
	// Tree is empty
	if this.root.length == 0 {
		return -1
	}

	node, index := this.Search(key)

	// Node not found
	if node == nil {
		return -1
	}

	node.Delete(index)

	return key
}
