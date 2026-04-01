package ssz

import "fmt"

type Wrapper struct {
	nodes []*Node
}

func (w *Wrapper) Indx() int {
	return len(w.nodes)
}

func (w *Wrapper) AddBytes(b []byte) {
	w.AddNode(LeafFromBytes(b))
}

// AddUint adds a uint8, uint16, uint32, or uint64 leaf node.
func AddUint[T marshalUints](w *Wrapper, i T) {
	w.AddNode(LeafFromUint(i))
}

// AddUint64 adds a uint64 leaf node.
//
// Deprecated: use AddUint instead.
func (w *Wrapper) AddUint64(i uint64) {
	AddUint(w, i)
}

// AddUint32 adds a uint32 leaf node.
//
// Deprecated: use AddUint instead.
func (w *Wrapper) AddUint32(i uint32) {
	AddUint(w, i)
}

// AddUint16 adds a uint16 leaf node.
//
// Deprecated: use AddUint instead.
func (w *Wrapper) AddUint16(i uint16) {
	AddUint(w, i)
}

// AddUint8 adds a uint8 leaf node.
//
// Deprecated: use AddUint instead.
func (w *Wrapper) AddUint8(i uint8) {
	AddUint(w, i)
}

func (w *Wrapper) AddNode(n *Node) {
	if w.nodes == nil {
		w.nodes = []*Node{}
	}
	w.nodes = append(w.nodes, n)
}

func (w *Wrapper) Node() *Node {
	if len(w.nodes) != 1 {
		fmt.Println(w.nodes)
		panic("BAD")
	}
	return w.nodes[0]
}

func (w *Wrapper) Commit(i int) {
	res, err := TreeFromNodes(w.nodes[i:])
	if err != nil {
		panic(err)
	}
	// remove the old nodes
	w.nodes = w.nodes[:i]
	// add the new node
	w.AddNode(res)
}

func (w *Wrapper) CommitWithMixin(i, num, limit int) {
	res, err := TreeFromNodesWithMixin(w.nodes[i:], num, limit)
	if err != nil {
		panic(err)
	}
	// remove the old nodes
	w.nodes = w.nodes[:i]
	// add the new node
	w.AddNode(res)
}

func (w *Wrapper) AddEmpty() {
	w.AddNode(EmptyLeaf())
}
