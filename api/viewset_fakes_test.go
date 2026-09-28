package api

import "context"

// Router.Register checks that a writable viewset's Queryset supports every
// action it routes. Fakes whose tests never reach an action embed these
// stubs to satisfy that check without changing what the tests exercise.

// unusedListOperations provides the Count and All methods list needs.
type unusedListOperations struct{}

func (unusedListOperations) Count(context.Context) (int64, error)       { return 0, nil }
func (unusedListOperations) All(context.Context) ([]interface{}, error) { return nil, nil }

// unusedUpdateOperation provides the Update method update needs.
type unusedUpdateOperation struct{}

func (unusedUpdateOperation) Update(context.Context, interface{}) error { return nil }

// unusedDeleteOperation provides the Delete method destroy needs.
type unusedDeleteOperation struct{}

func (unusedDeleteOperation) Delete(context.Context, interface{}) error { return nil }
