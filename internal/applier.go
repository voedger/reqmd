package internal

type dummyApplier struct{}

func NewDummyApplier() IApplier {
	return &dummyApplier{}
}

func (a *dummyApplier) Apply(ar *AnalyzerResult) error {
	return nil
}
