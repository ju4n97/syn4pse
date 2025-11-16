package mapsafe

// Get retrieves a typed value from a map[string]any.
// If the key is missing or the type cannot be converted, it returns the default value.
func Get[T any](m map[string]any, key string, defaultValue T) T {
	val, ok := m[key]
	if !ok {
		return defaultValue
	}

	if result, ok := val.(T); ok {
		return result
	}

	if result, ok := tryConvert[T](val); ok {
		return result
	}

	return defaultValue
}

// tryConvert attempts to convert val to type T with common numeric conversions
func tryConvert[T any](val any) (T, bool) {
	var zero T

	if _, isInt := any(zero).(int); isInt {
		switch v := val.(type) {
		case int:
			if result, ok := any(v).(T); ok {
				return result, true
			}
		case float64:
			if result, ok := any(int(v)).(T); ok {
				return result, true
			}
		}
	}

	if _, isFloat := any(zero).(float64); isFloat {
		switch v := val.(type) {
		case float64:
			if result, ok := any(v).(T); ok {
				return result, true
			}
		case int:
			if result, ok := any(float64(v)).(T); ok {
				return result, true
			}
		}
	}

	return zero, false
}
