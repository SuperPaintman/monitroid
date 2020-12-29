package gatherers

type Gatherer interface {
	Gather() (interface{}, error)
}
