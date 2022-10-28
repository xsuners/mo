package database

type Operation string

const (
	Operation_eq Operation = "eq"
	Operation_ne Operation = "ne"
	Operation_ge Operation = "ge"
	Operation_gt Operation = "gt"
	Operation_le Operation = "le"
	Operation_lt Operation = "lt"
	Operation_in Operation = "in"
)

type Condition struct {
	Op    Operation
	Key   string
	Value any
}

type Q func(func(*Condition))

func EQ(key string, value any) Q {
	return func(f func(*Condition)) {
		f(&Condition{
			Op:    Operation_eq,
			Key:   key,
			Value: value,
		})
	}
}

func GT(key string, value any) Q {
	return func(f func(*Condition)) {
		f(&Condition{
			Op:    Operation_gt,
			Key:   key,
			Value: value,
		})
	}
}

func GE(key string, value any) Q {
	return func(f func(*Condition)) {
		f(&Condition{
			Op:    Operation_ge,
			Key:   key,
			Value: value,
		})
	}
}

func IN(key string, value any) Q {
	return func(f func(*Condition)) {
		f(&Condition{
			Op:    Operation_in,
			Key:   key,
			Value: value,
		})
	}
}
