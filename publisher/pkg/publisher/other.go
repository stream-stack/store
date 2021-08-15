package publisher

import "fmt"

type SystemStreamOverview struct {
	Points map[string]UIntStartPoint `json:"points"`
}

func (o *SystemStreamOverview) minPoint() StartPoint {
	var minPoint StartPoint
	for _, point := range o.Points {
		if minPoint == nil {
			minPoint = point
			continue
		}
		if point.Compare(minPoint) > 0 {
			minPoint = point
		}
	}
	return minPoint
}

type UIntStartPoint struct {
	Point uint64 `json:"point"`
}

func (u UIntStartPoint) Compare(s StartPoint) int {
	point, ok := s.(UIntStartPoint)
	if !ok {
		panic(fmt.Errorf("startpoint %+v not type UIntStartPoint", s))
	}
	if u.Point > point.Point {
		return 1
	}
	if u.Point < point.Point {
		return -1
	}
	return 0
}
