package runtime

type StageCoordinator struct {
	runners []*StageRunner
}

func NewStageCoordinator(runners ...*StageRunner) *StageCoordinator {
	return &StageCoordinator{runners: runners}
}
