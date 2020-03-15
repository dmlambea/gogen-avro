package parser

func stringFromMap(m map[string]interface{}, key string) (string, error) {
	val, ok := m[key]
	if !ok {
		return "", NewRequiredMapKeyError(key)
	}
	if typedVal, ok := val.(string); ok {
		return typedVal, nil
	}
	return "", NewWrongMapValueTypeError(key, "string", val)
}

func arrayFromMap(m map[string]interface{}, key string) ([]interface{}, error) {
	val, ok := m[key]
	if !ok {
		return nil, NewRequiredMapKeyError(key)
	}
	if typedVal, ok := val.([]interface{}); ok {
		return typedVal, nil
	}
	return nil, NewWrongMapValueTypeError(key, "array", val)
}

func floatFromMap(m map[string]interface{}, key string) (float64, error) {
	val, ok := m[key]
	if !ok {
		return 0, NewRequiredMapKeyError(key)
	}
	if typedVal, ok := val.(float64); ok {
		return typedVal, nil
	}
	return 0, NewWrongMapValueTypeError(key, "float", val)
}
