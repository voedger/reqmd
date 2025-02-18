package internal

import "log"

type dummyApplier struct{}

func NewDummyApplier() IApplier {
	return &dummyApplier{}
}

func (a *dummyApplier) Apply(actions []Action) error {
	for _, action := range actions {
		switch action.Type {
		case ActionAnnotate:
			log.Printf("[DUMMY] Would update content at %s:%d: %s\n",
				action.Path, action.Line, action.Data)
		default:
			log.Printf("[DUMMY] Unknown action type: %v\n", action.Type)
		}
	}
	return nil
}
