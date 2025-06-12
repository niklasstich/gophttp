package common

import (
	"fmt"
	"strings"
)

type RadixTree[T any] struct {
	Node *RadixTreeNode[T]
}

func NewRadixTree[T any]() *RadixTree[T] {
	return &RadixTree[T]{&RadixTreeNode[T]{Children: make([]*RadixTreeEdge[T], 0)}}
}

type Radix[T any] interface {
	Find(path string) (T, error)
	Insert(path string, data T) error
	Delete(path string) error
	Nodes() int
}

var ErrNoMatch = fmt.Errorf("no match found")
var ErrPathAlreadyExists = fmt.Errorf("path already exists")

func (r RadixTree[T]) Find(path string) (T, error) {
	currNode := r.Node
	for len(path) > 0 {
		var err error
		currNode, path, err = findNextNode(currNode, path)
		if err != nil {
			return *new(T), err
		}
	}
	if currNode.HasData {
		return currNode.Data, nil
	}
	return *new(T), ErrNoMatch
}

func findNextNode[T any](currNode *RadixTreeNode[T], path string) (*RadixTreeNode[T], string, error) {
	for _, child := range currNode.Children {
		unmatched := child.Label.Matches(path)
		if unmatched != path {
			return child.Node, unmatched, nil
		}
	}
	//no match found in tree
	return nil, "", ErrNoMatch
}

func (r RadixTree[T]) Insert(path string, data T) error {
	//first, try to find the deepest node that matches our path
	currNode := r.Node
	for len(path) > 0 {
		var err error
		//preserve currNode in case we error out
		currNodeTemp := currNode
		pathTemp := path
		currNode, path, err = findNextNode(currNode, path)
		if err != nil {
			currNode = currNodeTemp
			path = pathTemp
			//can't find anything that matches anymore, so we found the deepest node
			break
		}
	}
	if len(path) == 0 {
		if currNode.HasData {
			return ErrPathAlreadyExists
		} else {
			//simply add data to existing node
			currNode.Data = data
			currNode.HasData = true
			return nil
		}
	}
	//if children exist, see if any child has a matching prefix and find the longest prefix that still matches
	if len(currNode.Children) > 0 {
		for _, child := range currNode.Children {
			//skip all edges that are variables
			e, ok := child.Label.(RadixTreeStringLabel)
			if !ok {
				continue
			}
			matchedPrefix := LongestCommonPrefix(e.Label, path)
			if len(matchedPrefix) == 0 {
				continue
			}
			//we found a match, edit current edge to be the prefix and add two new nodes for the suffixes
			existingNodeSuffix := strings.TrimPrefix(e.Label, matchedPrefix)
			newNodeSuffix := strings.TrimPrefix(path, matchedPrefix)
			if len(newNodeSuffix) == 0 {
				//existing node will now have data instead of creating an empty "" label edge
				existingNodeReplacement := RadixTreeNode[T]{child.Node.Data, child.Node.HasData, child.Node.Children}
				existingEdge := RadixTreeEdge[T]{RadixTreeStringLabel{existingNodeSuffix}, &existingNodeReplacement}
				newChildren := []*RadixTreeEdge[T]{&existingEdge}

				child.Node.HasData = true
				child.Node.Data = data
				child.Node.Children = newChildren
				//we have to also change the label
				child.Label = RadixTreeStringLabel{matchedPrefix}
				return nil
			} else {
				//create replacement for current child which will become child of a new node
				existingNodeReplacement := RadixTreeNode[T]{child.Node.Data, child.Node.HasData, child.Node.Children}
				//create the new node itself
				newNode := RadixTreeNode[T]{data, true, []*RadixTreeEdge[T]{}}
				//override current edge with a new edge
				existingEdge := RadixTreeEdge[T]{RadixTreeStringLabel{existingNodeSuffix}, &existingNodeReplacement}
				newEdge := RadixTreeEdge[T]{RadixTreeStringLabel{newNodeSuffix}, &newNode}
				newChildren := []*RadixTreeEdge[T]{&existingEdge, &newEdge}

				child.Node.HasData = false
				child.Node.Children = newChildren
				child.Label = RadixTreeStringLabel{matchedPrefix}
				return nil
			}

		}
	}

	//there are either no children, or no child that has a matching prefix
	//create a new edge from scratch
	newNode := RadixTreeNode[T]{data, true, []*RadixTreeEdge[T]{}}
	newEdge := RadixTreeEdge[T]{RadixTreeStringLabel{path}, &newNode}
	currNode.Children = append(currNode.Children, &newEdge)
	return nil
}

func (r RadixTree[T]) Delete(path string) error {
	//TODO implement me
	panic("implement me")
}

func (r RadixTree[T]) Nodes() int {
	currNode := r.Node
	count := 0
	nodesToCheck := make([]*RadixTreeNode[T], 0)
	for {
		count += len(currNode.Children)
		for _, child := range currNode.Children {
			nodesToCheck = append(nodesToCheck, child.Node)
		}
		if len(nodesToCheck) == 1 {
			currNode, nodesToCheck = nodesToCheck[0], make([]*RadixTreeNode[T], 0)
			continue
		}
		if len(nodesToCheck) > 0 {
			currNode, nodesToCheck = nodesToCheck[0], nodesToCheck[1:]
			continue
		}
		break
	}
	return count
}

type RadixTreeEdge[T any] struct {
	Label RadixTreeLabel
	Node  *RadixTreeNode[T]
}

type RadixTreeNode[T any] struct {
	Data     T
	HasData  bool
	Children []*RadixTreeEdge[T]
}

type RadixTreeLabel interface {
	Matches(path string) string
}

type RadixTreeStringLabel struct {
	Label string
}

type RadixTreeVariableLabel struct {
	VariableName string
}

func (sl RadixTreeStringLabel) Matches(path string) string {
	//return sl.Label == path
	if strings.HasPrefix(path, sl.Label) {
		return path[len(sl.Label):]
	}
	return path
}

// Matches for a variable label should match anything up to the next slash (or end of stream)
func (vl RadixTreeVariableLabel) Matches(path string) string {
	splits := strings.SplitN(path, "/", 2)
	if len(splits) == 1 {
		return ""
	} else {
		return splits[1]
	}
}
