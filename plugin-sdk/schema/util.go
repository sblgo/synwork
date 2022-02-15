package schema

import "strconv"

func MapCopy(m map[string]interface{}) map[string]interface{} {
	n := map[string]interface{}{}
	for k, v := range m {
		switch t := v.(type) {
		//		case *Reference:
		case []interface{}:
			n[k] = ListCopy(t)
		case map[string]interface{}:
			t2 := MapCopy(t)
			n[k] = t2
		default:
			n[k] = t
		}
	}
	return n
}

func ListCopy(l []interface{}) []interface{} {
	n := []interface{}{}
	for _, e := range l {
		switch t := e.(type) {
		case map[string]interface{}:
			n = append(n, MapCopy(t))
		case []interface{}:
			n = append(n, ListCopy(t))
		default:
			n = append(n, t)
		}
	}
	return n
}

func GetValueMap(m interface{}, path []string) (interface{}, bool) {
	if len(path) == 0 {
		if m != nil {
			return m, true
		} else {
			return nil, false
		}
	}
	switch t := m.(type) {
	case []interface{}:
		if len(path) > 1 {
			if i, err := strconv.Atoi(path[1]); err != nil {
				if 0 <= i && i < len(t) {
					return GetValueMap(t[i], path[1:])
				} else {
					return nil, false
				}
			} else {
				panic("invalid path")
			}
		} else {
			return t, true
		}
	case map[string]interface{}:
		return GetValueMap(t[path[0]], path[1:])
	}
	return nil, false
}

func SetValueMap(m map[string]interface{}, path []string, v interface{}) bool {
	if len(path) == 1 {
		m[path[0]] = v
		return true
	} else if len(path) == 0 {
		return false
	}
	if s, ok := m[path[0]]; ok {
		switch t := s.(type) {
		case map[string]interface{}:
			return SetValueMap(t, path[1:], v)
		case []interface{}:
			if pos, err := strconv.Atoi(path[1]); err != nil {
				return false
			} else if 0 <= pos && pos < len(t) && len(path) == 2 {
				t[pos] = v
				m[path[0]] = t
				return true
			} else if 0 <= pos && pos < len(t) && len(path) > 2 {
				nv := t[pos]
				switch nvt := nv.(type) {
				case map[string]interface{}:
					return SetValueMap(nvt, path[2:], v)
				default:
					return false
				}
			} else if pos == len(t) && len(path) == 2 {
				t = append(t, v)
				m[path[0]] = t
				return true
			} else if pos == len(t) && len(path) > 2 {
				sv := map[string]interface{}{}
				t = append(t, sv)
				m[path[0]] = t
				return SetValueMap(sv, path[2:], v)
			} else {
				return false
			}
		default:
			return false
		}
	} else {

	}
	return false
}
