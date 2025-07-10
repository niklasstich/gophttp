package common

import (
	"testing"
)

func TestRadixTree_Find(t *testing.T) {
	// Create a RadixTree with some nodes
	tree := RadixTree[int]{
		Node: &RadixTreeNode[int]{
			Data:    0,
			HasData: false,
			Children: []*RadixTreeEdge[int]{
				{
					Label: RadixTreeStringLabel{Label: "home"},
					Node: &RadixTreeNode[int]{
						Data:    1,
						HasData: true,
						Children: []*RadixTreeEdge[int]{
							{
								Label: RadixTreeStringLabel{Label: "/"},
								Node: &RadixTreeNode[int]{
									Data: 0,
									Children: []*RadixTreeEdge[int]{
										{
											Label: RadixTreeStringLabel{Label: "about"},
											Node: &RadixTreeNode[int]{
												Data:    2,
												HasData: true,
											},
										},
										{
											Label: RadixTreeStringLabel{Label: "contact"},
											Node: &RadixTreeNode[int]{
												Data:    3,
												HasData: true,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Label: RadixTreeStringLabel{Label: "api"},
					Node: &RadixTreeNode[int]{
						HasData: false,
						Children: []*RadixTreeEdge[int]{
							{
								Label: RadixTreeStringLabel{Label: "/"},
								Node: &RadixTreeNode[int]{
									HasData: false,
									Children: []*RadixTreeEdge[int]{
										{
											Label: RadixTreeVariableLabel{VariableName: "resource"},
											Node: &RadixTreeNode[int]{
												Data:    5,
												HasData: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Test cases
	tests := []struct {
		path     string
		expected int
	}{
		{"home", 1},
		{"home/", -1},
		{"home/about", 2},
		{"home/contact", 3},
		{"home/contact/", -1},
		{"api", -1},
		{"api/", -1},
		{"api/users", 5},
		{"api/products", 5},
		{"api/endwithslash/", 5},
		{"unknown", -1},
		{"", -1},
	}

	for _, test := range tests {
		node, err := tree.Find(test.path)
		if err != nil {
			if test.expected != -1 {
				t.Errorf("unexpected error for path %s: %v", test.path, err)
			}
		} else if node != test.expected {
			t.Errorf("expected %d for path %s, got %d", test.expected, test.path, node)
		}
	}
}

func TestRadixTree_Insert(t *testing.T) {
	tree := RadixTree[int]{Node: &RadixTreeNode[int]{0, false, []*RadixTreeEdge[int]{}}}

	tests := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{
		{path: "home/", data: 1, expected: nil, nodes: 1},                  // Insert into empty tree
		{path: "home/about/", data: 2, expected: nil, nodes: 2},            // Insert with shared prefix
		{path: "home/contact/", data: 3, expected: nil, nodes: 3},          // Insert another branch
		{path: "home/", data: 1, expected: ErrPathAlreadyExists, nodes: 3}, // Insert duplicate path
		{path: "api/users/", data: 4, expected: nil, nodes: 4},             // Insert new branch
		{path: "api/products/", data: 5, expected: nil, nodes: 6},          // Insert with shared prefix
		{path: "api/", data: 6, expected: nil, nodes: 6},                   // Insert shorter prefix
	}

	toCheck := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{}

	for _, test := range tests {
		err := tree.Insert(test.path, test.data)
		if err != test.expected {
			t.Errorf("Insert(%q, %d): expected error %v, got %v", test.path, test.data, test.expected, err)
		}
		if test.expected == nil {
			toCheck = append(toCheck, test)
		}

		// Verify lookup of all elements inserted so far
		for _, test := range toCheck {
			node, err := tree.Find(test.path)
			if err != nil {
				t.Errorf("Find(%q): unexpected error after insertion: %v", test.path, err)
			} else if node != test.data {
				t.Errorf("Find(%q): expected data %d, got %d", test.path, test.data, node)
			}
		}
		//assert tree has correct amount of nodes
		if tree.Nodes() != test.nodes {
			t.Errorf("expected %d nodes, found %d", test.nodes, tree.Nodes())
		}
	}
}

func TestRadixTree_InsertSubpathAfterFullPath(t *testing.T) {
	tree := RadixTree[int]{Node: &RadixTreeNode[int]{0, false, []*RadixTreeEdge[int]{}}}

	tests := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{
		{path: "home/about/", data: 2, expected: nil, nodes: 1},
		{path: "home/contact/", data: 3, expected: nil, nodes: 3},
		{path: "home/", data: 1, expected: nil, nodes: 3},
		{path: "home/", data: 1, expected: ErrPathAlreadyExists, nodes: 3},
		{path: "api/users/", data: 4, expected: nil, nodes: 4},
		{path: "api/products/", data: 5, expected: nil, nodes: 6},
		{path: "api/", data: 6, expected: nil, nodes: 6},
	}

	toCheck := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{}

	for _, test := range tests {
		err := tree.Insert(test.path, test.data)
		if err != test.expected {
			t.Errorf("Insert(%q, %d): expected error %v, got %v", test.path, test.data, test.expected, err)
		}
		if test.expected == nil {
			toCheck = append(toCheck, test)
		}

		// Verify lookup of all elements inserted so far
		for _, test := range toCheck {
			node, err := tree.Find(test.path)
			if err != nil {
				t.Errorf("Find(%q): unexpected error after insertion: %v", test.path, err)
			} else if node != test.data {
				t.Errorf("Find(%q): expected data %d, got %d", test.path, test.data, node)
			}
		}
		//assert tree has correct amount of nodes
		if tree.Nodes() != test.nodes {
			t.Errorf("expected %d nodes, found %d", test.nodes, tree.Nodes())
		}
	}
}

func TestRadixTree_InsertCreatesNoEdgeWithEmptyString(t *testing.T) {
	tree := RadixTree[int]{Node: &RadixTreeNode[int]{0, false, []*RadixTreeEdge[int]{}}}

	tests := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{
		{path: "home/about/", data: 2, expected: nil, nodes: 1},
		{path: "home/", data: 1, expected: nil, nodes: 2},
	}

	toCheck := []struct {
		path     string
		data     int
		expected error
		nodes    int
	}{}

	for _, test := range tests {
		err := tree.Insert(test.path, test.data)
		if err != test.expected {
			t.Errorf("Insert(%q, %d): expected error %v, got %v", test.path, test.data, test.expected, err)
		}
		if test.expected == nil {
			toCheck = append(toCheck, test)
		}

		// Verify lookup of all elements inserted so far
		for _, test := range toCheck {
			node, err := tree.Find(test.path)
			if err != nil {
				t.Errorf("Find(%q): unexpected error after insertion: %v", test.path, err)
			} else if node != test.data {
				t.Errorf("Find(%q): expected data %d, got %d", test.path, test.data, node)
			}
		}
		//assert tree has correct amount of nodes
		if tree.Nodes() != test.nodes {
			t.Errorf("expected %d nodes, found %d", test.nodes, tree.Nodes())
		}
	}
}
