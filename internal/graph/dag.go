package graph

import (
	"fmt"
	"hl/ui"
)

type Node struct {
	Step     ui.Step
	Children []*Node
	Parents  []*Node
	InDegree int
}

type Graph struct {
	Nodes map[string]*Node
}

func generateNodes(steps []ui.Step) (map[string]*Node, error) {
	nodes := make(map[string]*Node, len(steps))

	for _, step := range steps {
		if _, exists := nodes[step.Name]; exists {
			return nil, fmt.Errorf("duplicate step name detected: %q", step.Name)
		}

		nodes[step.Name] = &Node{
			Step:     step,
			Children: []*Node{},
			Parents:  []*Node{},
			InDegree: 0,
		}
	}

	return nodes, nil
}

func GenerateDAG(pipeline ui.Pipeline) (*Graph, error) {
	nodes, err := generateNodes(pipeline.Steps)
	if err != nil {
		return nil, err
	}

	g := &Graph{Nodes: nodes}

	for _, step := range pipeline.Steps {
		currentNode := g.Nodes[step.Name]

		for _, parentName := range step.DependsOn {
			parentNode, exists := g.Nodes[parentName]
			if !exists {
				return nil, fmt.Errorf("step %q depends on non-existent step %q", step.Name, parentName)
			}

			if parentName == step.Name {
				return nil, fmt.Errorf("step %q cannot depend on itself", step.Name)
			}

			parentNode.Children = append(parentNode.Children, currentNode)
			currentNode.Parents = append(currentNode.Parents, parentNode)
			currentNode.InDegree++
		}
	}

	return g, nil
}
