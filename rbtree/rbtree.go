package rbtree

import (
	"fmt"
	"os"
)

type NodeValue interface {
	Less(node NodeValue) bool
}

const (
	NODE_COLOR_RED   = iota
	NODE_COLOR_BLACK = 1
)

type rbnode struct {
	value  NodeValue
	color  uint8
	left   *rbnode
	right  *rbnode
	parent *rbnode
}

var nil_node = &rbnode{
	color: NODE_COLOR_BLACK,
}

func (this *rbnode) is_root() bool {
	return this.parent == nil
}

func (this *rbnode) is_nil() bool {
	return this.value == nil && this.color == NODE_COLOR_BLACK
}

func (this *rbnode) color_is_red() bool {
	return this.color == NODE_COLOR_RED
}

func (this *rbnode) color_is_black() bool {
	return this.color == NODE_COLOR_BLACK
}

func (this *rbnode) color_set_red() {
	this.color = NODE_COLOR_RED
}

func (this *rbnode) color_set_black() {
	this.color = NODE_COLOR_BLACK
}

func (this *rbnode) get_uncle() *rbnode {
	if this.parent == nil || this.parent.parent == nil {
		return nil
	}
	grandparent := this.parent.parent
	if grandparent.left == this.parent {
		return grandparent.right
	}
	return grandparent.left
}

func (this *rbnode) get_brother() *rbnode {
	if this.parent == nil {
		return nil
	}
	if this.parent.left == this {
		return this.parent.right
	}
	return this.parent.left
}

func (this *rbnode) get_grandparent() *rbnode {
	if this.parent == nil {
		return nil
	}
	return this.parent.parent
}

type RBTree struct {
	root     *rbnode
	node_num uint32
}

func is_nil(node *rbnode) bool {
	if node == nil || node.is_nil() {
		return true
	}
	return false
}

func (this *RBTree) _get_insert_parent(value NodeValue) (is_equal bool, node *rbnode, left_or_right bool) {
	if this.root == nil {
		return
	}

	node = this.root
	for {
		if value.Less(node.value) {
			// leaf
			if is_nil(node.left) {
				left_or_right = true
				break
			}
			node = node.left
		} else if node.value.Less(value) {
			// leaf
			if is_nil(node.right) {
				break
			}
			node = node.right
		} else {
			is_equal = true
			break
		}
	}
	return
}

func (this *RBTree) rotate_left(node *rbnode) bool {
	if is_nil(node) {
		return false
	}

	node_right := node.right
	if is_nil(node_right) {
		return false
	}

	parent := node.parent
	if parent != nil {
		if node == parent.left {
			parent.left = node_right
		} else {
			parent.right = node_right
		}
	}

	node_right.parent = parent
	node.parent = node_right
	right_left := node_right.left
	node_right.left = node
	node.right = right_left
	right_left.parent = node

	if node == this.root {
		this.root = node_right
	}

	return true
}

func (this *RBTree) rotate_right(node *rbnode) bool {
	if is_nil(node) {
		return false
	}

	node_left := node.left
	if is_nil(node_left) {
		return false
	}

	parent := node.parent
	if parent != nil {
		if node == parent.left {
			parent.left = node_left
		} else {
			parent.right = node_left
		}
	}

	node_left.parent = parent
	node.parent = node_left
	left_right := node_left.right
	node_left.right = node
	node.left = left_right
	left_right.parent = node

	if node == this.root {
		this.root = node_left
	}

	return true
}

func (this *RBTree) insert_fixup(node *rbnode, left_or_right bool) {
	for {
		parent := node.parent
		if parent == nil || parent.color_is_black() {
			break
		}
		uncle := node.get_uncle()
		grandparent := node.get_grandparent()
		if uncle.color_is_red() { // 叔父节点是红色
			// 变色
			parent.color_set_black()
			uncle.color_set_black()
			grandparent.color_set_red()
			node = grandparent
			continue
		}

		if parent == grandparent.left { // 父节点是祖父节点的左节点
			if !left_or_right { // 插入节点是父节点的右子节点
				// 左旋
				if !this.rotate_left(parent) {
					return
				}
				node, parent = parent, node
			}
			// 变色
			parent.color_set_black()
			grandparent.color_set_red()
			// 右旋
			if !this.rotate_right(grandparent) {
				return
			}
		} else { // 父节点是祖父节点的右节点
			if left_or_right { // 插入节点是父节点的左子节点
				// 右旋
				if !this.rotate_right(parent) {
					return
				}
				node, parent = parent, node
			}
			// 变色
			parent.color_set_black()
			grandparent.color_set_red()
			// 左旋
			if !this.rotate_left(grandparent) {
				return
			}
		}
	}
	this.root.color_set_black()
}

func (this *RBTree) Insert(value NodeValue) {
	is_equal, insert_parent, left_or_right := this._get_insert_parent(value)
	if is_equal {
		insert_parent.value = value
		return
	}

	// new node is red
	node := &rbnode{
		value:  value,
		color:  NODE_COLOR_RED,
		left:   nil_node,
		right:  nil_node,
		parent: insert_parent,
	}

	this.node_num += 1

	// empty tree
	if insert_parent == nil {
		node.color_set_black()
		this.root = node
		return
	}

	// insert position
	if left_or_right {
		insert_parent.left = node
	} else {
		insert_parent.right = node
	}

	this.insert_fixup(node, left_or_right)

	return
}

func (this *RBTree) get_substitude(node *rbnode) *rbnode {
	// 先找后继
	tmp := node.right
	for !is_nil(tmp) {
		tmp = tmp.left
	}

	if is_nil(tmp) {
		// 再找前驱
		tmp = node.left
		for !is_nil(tmp) {
			tmp = tmp.right
		}
	}

	return tmp
}

func (this *RBTree) Delete(value NodeValue) bool {
	is_equal, node, _ := this._get_insert_parent(value)
	if !is_equal {
		return false
	}

	has_left := node.left != nil && !node.left.is_nil()
	has_right := node.right != nil && !node.right.is_nil()
	// 被删除节点的左右孩子节点都是nil
	if !has_left && !has_right {
		if node.color == NODE_COLOR_RED { // 被删除的是红色节点
			if node == node.parent.left {
				node.parent.left = nil_node
			} else {
				node.parent.right = nil_node
			}
		} else { // 被删除的是黑色节点
			bnode := node.get_brother()
			if bnode == nil || bnode.is_nil() {
				fmt.Fprintf(os.Stderr, "rbtree delete black node no brother node")
				return false
			}
			b_has_left := bnode.left != nil && !bnode.left.is_nil()
			b_has_right := bnode.right != nil && !bnode.right.is_nil()
			// 兄弟节点的两个子节点都为nil
			if !b_has_left && !b_has_right {
				node.parent.color = NODE_COLOR_BLACK
				bnode.color = NODE_COLOR_RED
			}
		}
	}

	return true
}

func (this *RBTree) NodeNum() uint32 {
	return this.node_num
}

type stack struct {
	node_list []*rbnode
	top       int
}

func new_stack(length int) *stack {
	return &stack{
		node_list: make([]*rbnode, length),
	}
}

func (this *stack) is_empty() bool {
	return this.top == 0
}

func (this *stack) is_full() bool {
	return this.top == len(this.node_list)
}

func (this *stack) get_top() *rbnode {
	return this.node_list[this.top-1]
}

func (this *stack) push(node *rbnode) bool {
	if this.is_full() {
		return false
	}

	this.node_list[this.top] = node
	this.top += 1

	return true
}

func (this *stack) pop() *rbnode {
	if this.is_empty() {
		return nil
	}
	this.top -= 1
	return this.node_list[this.top]
}

func (this *RBTree) PreorderTraverse(f func(value NodeValue)) {
	p := this.root
	if p == nil {
		return
	}
	s := new_stack(int(this.node_num))
	s.push(p)
	for !s.is_empty() {
		p = s.pop()
		if p == nil || p.is_nil() {
			break
		}
		if f != nil {
			f(p.value)
		}
		if p.right != nil && !p.right.is_nil() {
			s.push(p.right)
		}
		if p.left != nil && !p.left.is_nil() {
			s.push(p.left)
		}
	}
}

func (this *RBTree) InorderTraverse(f func(value NodeValue)) {
	p := this.root
	if p == nil {
		return
	}
	s := new_stack(int(this.node_num))
	for !s.is_empty() || (p != nil && !p.is_nil()) {
		for p != nil && !p.is_nil() {
			s.push(p)
			p = p.left
		}
		p = s.pop()
		if f != nil {
			f(p.value)
		}
		p = p.right
	}
}

func (this *RBTree) PostorderTraverse(f func(value NodeValue)) {
	cur := this.root
	if cur == nil {
		return
	}
	var pre *rbnode
	s := new_stack(int(this.node_num))
	for (cur != nil && !cur.is_nil()) || !s.is_empty() {
		// 把左子节点依次入栈直到叶节点
		for cur != nil && !cur.is_nil() {
			s.push(cur)
			cur = cur.left
		}
		cur = s.get_top()
		if cur.right == nil || cur.right.is_nil() || cur.right == pre {
			s.pop()
			f(cur.value)
			pre = cur
			cur = nil
		} else {
			cur = cur.right
		}
	}
}

type queue struct {
	nodes       []*rbnode
	write_index int
	node_num    int
}

func new_queue(size int) *queue {
	return &queue{
		nodes: make([]*rbnode, size),
	}
}

func (this *queue) num() int {
	return this.node_num
}

func (this *queue) is_empty() bool {
	return this.node_num == 0
}

func (this *queue) get_front() *rbnode {
	if this.node_num == 0 {
		return nil
	}
	read_index := this.write_index - this.node_num
	if read_index < 0 {
		read_index += len(this.nodes)
	}
	return this.nodes[read_index]
}

func (this *queue) push_back(node *rbnode) bool {
	if this.node_num >= len(this.nodes) {
		return false
	}

	this.nodes[this.write_index] = node
	this.write_index += 1
	this.node_num += 1

	return true
}

func (this *queue) pop_front() bool {
	if this.node_num == 0 {
		return false
	}
	this.node_num -= 1
	return true
}

func (this *RBTree) BFSTraverse(f func(value NodeValue)) {
	cur := this.root
	if cur == nil {
		return
	}
	q := new_queue(int(this.node_num))
	q.push_back(cur)
	for !q.is_empty() {
		cur = q.get_front()
		q.pop_front()
		if cur == nil || !cur.is_nil() {
			break
		}
		f(cur.value)
		if cur.left != nil && !cur.left.is_nil() {
			q.push_back(cur.left)
		}
		if cur.right != nil && !cur.right.is_nil() {
			q.push_back(cur.right)
		}
	}
}

func (this *RBTree) DFSTraverse(f func(value NodeValue)) {
	if this.root == nil {
		return
	}
	s := new_stack(int(this.node_num))
	s.push(this.root)
	var cur *rbnode
	for !s.is_empty() {
		cur = s.get_top()
		if cur == nil || !cur.is_nil() {
			break
		}
		s.pop()
		f(cur.value)
		if cur.right != nil && !cur.right.is_nil() {
			s.push(cur.right)
		}
		if cur.left != nil && !cur.left.is_nil() {
			s.push(cur.left)
		}
	}
}
