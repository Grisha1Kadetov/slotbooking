package log

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Close()
}

type Field struct {
	Key   string
	Value any
}

func Pair(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Err(value any) Field {
	return Field{Key: "error", Value: value}
}

func FromMap(fields map[string]any) []Field {
	result := make([]Field, 0, len(fields))
	for k, v := range fields {
		result = append(result, Pair(k, v))
	}
	return result
}
