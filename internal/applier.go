package internal

type applier struct{}

func NewApplier() IApplier {
	return &applier{}
}

func (a *applier) Apply(ar *AnalyzerResult) error {

	if IsVerbose {
		for _, actions := range ar.MdActions {
			for _, action := range actions {
				Verbose("Action\n\t" + action.String())
			}
		}
	}
	return nil
}
