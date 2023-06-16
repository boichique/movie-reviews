package dbx

type ChangeRelationFunc[S any] func(S) error

type Keyer interface {
	Key() any
}

func AdjustRelations[S interface {
	Keyer
	comparable
}](
	prev, next []S,
	addFn, removeFn ChangeRelationFunc[S],
) error {
	prevM, nextM := toMap(prev), toMap(next)
	for key, prevItem := range prevM {
		if nextItem, contains := nextM[key]; !contains || prevItem != nextItem {
			err := removeFn(prevItem)
			if err != nil {
				return err
			}
		}
	}

	for key, nextItem := range nextM {
		if prevItem, contains := prevM[key]; !contains || prevItem != nextItem {
			err := addFn(nextItem)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func toMap[S Keyer](slice []S) map[any]S {
	m := make(map[any]S, len(slice))
	for _, item := range slice {
		m[item.Key()] = item
	}

	return m
}
