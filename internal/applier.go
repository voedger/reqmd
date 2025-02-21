package internal

type applier struct {
	dryRun bool
}

func NewApplier(dryRun bool) IApplier {
	return &applier{
		dryRun: dryRun,
	}
}

func (a *applier) Apply(ar *AnalyzerResult) error {
	if a.dryRun || IsVerbose {
		Verbose("Actions that would be applied:")
		for _, actions := range ar.MdActions {
			for _, action := range actions {
				Verbose("Action\n\t" + action.String())
			}
		}
		if a.dryRun {
			return nil
		}
	}
	if a.dryRun {
		return nil
	}

	for path, actions := range ar.MdActions {
		err := applyActions(path, actions)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyActions(_ FilePath, _ []MdAction) error {
	return nil
}
