package tlb_pretty

import (
	"errors"
	"fmt"
)

const (
	AhmeRoot = "ahme_root"
	AhmnFork = "ahmn_fork"
	AhmEdge  = "ahm_edge"
	AhmnLeaf = "ahmn_leaf"

	HmeRoot = "hme_root"
	HmEdge  = "hm_edge"
	HmnFork = "hmn_fork"
	HmnLeaf = "hmn_leaf"

	BtFork = "bt_fork"
	BtLeaf = "bt_leaf"
)

type TreeSimplifier struct {
}

// Extract leafs from ahme and hme trees and remove edges.
// Like from:
// a -> b -> c -> d
//      |      \
//      f     	 > e
//
// to:
// a -> [d, e, f]
func (t *TreeSimplifier) Simplify(node *AstNode) (*AstNode, error) {
	simplifiedNode := &AstNode{
		Parent: node.Parent,
		Fields: make(map[string]interface{}),
	}

	if node.IsType(AhmeRoot) {
		leafs, err := t.extractLeafsRoot(node, AhmeRoot, AhmEdge, AhmnFork, AhmnLeaf)
		if err != nil {
			return nil, err
		}

		simplifiedNode.Fields["@type"] = AhmeRoot
		simplifiedNode.Fields["leafs"] = leafs
	} else if node.IsType(HmeRoot) {
		leafs, err := t.extractLeafsRoot(node, HmeRoot, HmEdge, HmnFork, HmnLeaf)
		if err != nil {
			return nil, err
		}

		simplifiedNode.Fields["@type"] = HmeRoot
		simplifiedNode.Fields["leafs"] = leafs
	} else {
		for k, v := range node.Fields {
			switch v.(type) {
			case *AstNode:
				n, err := t.Simplify(v.(*AstNode))
				if err != nil {
					return nil, err
				}
				simplifiedNode.Fields[k] = n
			default:
				simplifiedNode.Fields[k] = v
			}
		}
	}

	return simplifiedNode, nil
}

func (t *TreeSimplifier) extractLeafsRoot(node *AstNode, rootType string, edgeType string, forkType string, leafType string) ([]*AstNode, error) {
	if !node.IsType(rootType) {
		fmt.Println("node is not" + rootType)
		return nil, errors.New("node is not" + rootType)
	}
	edgeNode, err := node.GetNode("root")
	if err != nil {
		return nil, err
	}

	if !edgeNode.IsType(edgeType) {
		return nil, err
	}

	return t.edgeExtractLeafs(edgeNode, edgeType, forkType, leafType)
}

func (t *TreeSimplifier) edgeExtractLeafs(node *AstNode, edgeType string, forkType string, leafType string) ([]*AstNode, error) {
	if !node.IsType(edgeType) {
		return nil, errors.New("node is not" + edgeType)
	}

	nodeField, err := node.GetNode("node")
	if err != nil {
		return nil, err
	}

	leafs := make([]*AstNode, 0)

	if nodeField.IsType(forkType) {
		if left, err := nodeField.GetNode("left"); err == nil {
			if left.IsType(edgeType) {
				leftLeafs, err := t.edgeExtractLeafs(left, edgeType, forkType, leafType)
				if err != nil {
					return nil, err
				}
				leafs = append(leafs, leftLeafs...)
			} else {
				fmt.Println("left is not edge!!!")
			}
		}
		if right, err := nodeField.GetNode("right"); err == nil {
			if right.IsType(edgeType) {
				rightLeafs, err := t.edgeExtractLeafs(right, edgeType, forkType, leafType)
				if err != nil {
					return nil, err
				}
				leafs = append(leafs, rightLeafs...)
			} else {
				fmt.Println("left is not edge!!!")
			}
		}
	} else if nodeField.IsType(leafType) {
		nodeSimplifiedField, err := t.Simplify(nodeField)
		if err != nil {
			return nil, err
		}

		// leaf value is bt tree with forks and leafs
		leafValueNode, err := nodeSimplifiedField.GetNode("value")
		if err == nil {
			if leafValueNode.IsType(BtFork) {
				leafsBtFork, err := t.extractLeafsFork(leafValueNode, BtFork, BtLeaf)
				if err != nil {
					return nil, err
				}

				leafs = append(leafs, leafsBtFork...)
			} else {
				leafs = append(leafs, nodeSimplifiedField)
			}
		} else {
			leafs = append(leafs, nodeSimplifiedField)
		}
	}

	return leafs, nil
}

func (t *TreeSimplifier) extractLeafsFork(node *AstNode, forkType string, leafType string) ([]*AstNode, error) {
	if !node.IsType(forkType) {
		return nil, errors.New("is not fork")
	}
	leafs := make([]*AstNode, 0)

	if left, err := node.GetNode("left"); err == nil {
		if left.IsType(forkType) {
			leftLeafs, err := t.extractLeafsFork(left, forkType, leafType)
			if err != nil {
				return nil, err
			}
			leafs = append(leafs, leftLeafs...)
		} else if left.IsType(leafType) {
			leafs = append(leafs, left)
		} else {
			fmt.Println("left is not fork and not leaf!!!")
		}
	}

	if right, err := node.GetNode("right"); err == nil {
		if right.IsType(forkType) {
			rightLeafs, err := t.extractLeafsFork(right, forkType, leafType)
			if err != nil {
				return nil, err
			}
			leafs = append(leafs, rightLeafs...)
		} else if right.IsType(leafType) {
			leafs = append(leafs, right)
		} else {
			fmt.Println("right is not fork and not leaf!!!")
		}
	}

	return leafs, nil
}

func NewTreeSimplifier() *TreeSimplifier {
	return &TreeSimplifier{}
}
