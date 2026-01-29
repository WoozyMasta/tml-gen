package main

import "strings"

// Node stores directory aggregation for grouping.
type Node struct {
	Parent   *Node            // parent node
	Children map[string]*Node // children nodes
	Name     string           // directory name
	Count    int              // number of objects in the node
}

// Rec binds a relative path to its directory node.
type Rec struct {
	DirNode *Node  // directory node
	RelPath string // relative to game root, with '/'
}

// newNode creates a new node.
func newNode(name string, parent *Node) *Node {
	return &Node{Name: name, Parent: parent, Children: make(map[string]*Node)}
}

// insert inserts a directory segment into a node.
func insert(root *Node, dirSegs []string) *Node {
	n := root
	n.Count++
	for _, s := range dirSegs {
		ch := n.Children[s]
		if ch == nil {
			ch = newNode(s, n)
			n.Children[s] = ch
		}
		n = ch
		n.Count++
	}
	return n
}

// pickGroup picks a group node based on the threshold.
func pickGroup(n *Node, thr int) *Node {
	cur := n
	// Stop climbing at level-2 to avoid grouping into top-level.
	for cur != nil && cur.Parent != nil && cur.Parent.Parent != nil && cur.Parent.Parent.Parent != nil && cur.Count < thr {
		cur = cur.Parent
	}

	// If we ended up at root (shouldn't, but guard), keep original
	if cur == nil || cur.Parent == nil {
		return n
	}

	return cur
}

// nodeKey generates a key for a node.
func nodeKey(n *Node) string {
	if n == nil || n.Parent == nil {
		return "root"
	}

	var parts []string
	for cur := n; cur != nil && cur.Parent != nil; cur = cur.Parent {
		parts = append(parts, cur.Name)
	}
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, "_")
}
