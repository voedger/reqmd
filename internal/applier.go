package internal

import "fmt"

type dummyApplier struct{}

func NewDummyApplier() IApplier {
	return &dummyApplier{}
}

func (a *dummyApplier) Apply(ar *AnalyzerResult) error {
	for path, actions := range ar.MdActions {
		for _, action := range actions {
			Verbose(fmt.Sprintf("Applying action %v", action), "path", path)
		}
	}
	return nil
}
