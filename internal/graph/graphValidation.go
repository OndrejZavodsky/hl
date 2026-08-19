package graph

import "fmt"

func (g *Graph) Validate() error {
	inDegrees := make(map[string]int, len(g.Nodes))
	var zeroInDegree []*Node

	for name, node := range g.Nodes {
		inDegrees[name] = node.InDegree
		if node.InDegree == 0 {
			zeroInDegree = append(zeroInDegree, node)
		}
	}

	visited := 0
	for len(zeroInDegree) > 0 {
		curr := zeroInDegree[0]
		zeroInDegree = zeroInDegree[1:]
		visited++

		for _, child := range curr.Children {
			inDegrees[child.Step.Name]--
			if inDegrees[child.Step.Name] == 0 {
				zeroInDegree = append(zeroInDegree, child)
			}
		}
	}

	if visited != len(g.Nodes) {
		return fmt.Errorf("circular dependency detected in pipeline")
	}

	return nil
}
