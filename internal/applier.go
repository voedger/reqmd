package internal

type dummyApplier struct{}

func NewDummyApplier() IApplier {
	return &dummyApplier{}
}

func (a *dummyApplier) Apply(ar *AnalyzerResult) error {
	for _, actions := range ar.MdActions {
		for _, action := range actions {
			Verbose(action.String())
		}
	}
	return nil
}
