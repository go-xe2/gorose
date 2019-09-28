package sqlw

import (
	"github.com/gohouse/t"
	"reflect"
	"strings"
)

type IToString interface {
	String() string
}

func whereString(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	return map[string]interface{}{
		"$string": s,
	}
}

func whereSliceValue(field string, value []interface{}) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	if len(value) == 1 {
		return map[string]interface{}{field: []interface{}{"=", value[0]}}
	}
	if len(value) > 1 {
		if s, ok := value[0].(string); !ok {
			return nil
		} else {
			return map[string]interface{}{field: []interface{}{s, value[1]}}
		}
	}
	return nil
}

func whereCondition(condition string, args []interface{}) map[string]interface{} {
	if condition == "" || len(args) == 0 {
		return nil
	}
	var results []interface{}
	for _, arg := range args {
		argType := reflect.TypeOf(arg).Kind()
		switch argType {
		case reflect.String:
			results = append(results, map[string]interface{}{"$string": arg.(string)})
			break
		case reflect.Map:
			m := t.New(arg).MapStringInterface()
			if m == nil || len(m) == 0 {
				continue
			}
			if len(m) > 1 {
				if mItems := whereMap(m); mItems != nil {
					results = append(results, map[string]interface{}{ condition: mItems })
				}
				continue
			}
			// 只有1个字段
			for fd, v := range m {
				if whereItem := whereP2(fd, v); whereItem != nil {
					results = append(results, whereItem)
				}
			}
			break
		case reflect.Slice:
			arr := arg.([]interface{})
			arrLen := len(arr)
			switch arrLen {
			case 3:
				fd, ok1 := arr[0].(string)
				cp, ok2 := arr[1].(string)
				if !ok1 || !ok2 {
					continue
				}
				if v := whereP3(fd, cp, arr[2]); v != nil {
					results = append(results, v)
				}
				break
			case 2:
				if s, ok := arr[0].(string); !ok {
					continue
				} else {
					if v := whereP2(s, arr[1]); v != nil {
						results = append(results, v)
					}
				}
				break
			case 1:
				if v := whereP1(arr[0]); v != nil {
					results = append(results, v)
				}
				break
			}
		}
	}
	if len(results) == 0 {
		return nil
	}
	return map[string]interface{}{ condition: results }
}


func whereP1(express interface{}) interface{} {
	expType := reflect.TypeOf(express).Kind()
	switch expType {
	case reflect.String:
		if v := whereString(express.(string)); v != nil {
			return v
		}
		break
	case reflect.Map:
		if v, ok := express.(map[string]interface{}); ok {
			return whereMap(v)
		}
		break
	case reflect.Slice:
		arr1, ok := express.([][]interface{})
		if !ok {
			arr := t.New(express).SliceInterface()
			if arr != nil {
				return whereCondition("$and", arr)
			}
			return nil
		}
		var results []interface{}
		for _, item := range arr1 {
			itemLen := len(item)
			switch itemLen {
			case 3:
				fd, ok1 := item[0].(string)
				cp, ok2 := item[1].(string)
				if ok1 && ok2 {
					if v := whereP3(fd, cp, item[2]); v != nil {
						results = append(results, v)
					}
				}
				break
			case 2:
				fd, ok := item[0].(string)
				if ok {
					v := whereP2(fd, item[1])
					if v == nil {
						continue
					}
					results = append(results, v)
				}
				break
			case 1:
				v := whereP1(item[0])
				if v == nil {
					continue
				}
				results = append(results, v)
				break
			}
		}
		return results
	}
	return nil
}

func whereP2(fieldOrCondition string, value interface{}) interface{} {
	condition := strings.ToLower(fieldOrCondition)
	isCondition := condition == "$or" || condition == "$and"
	vType := reflect.TypeOf(value).Kind()
	if isCondition && vType != reflect.Slice && vType != reflect.Map {
		return nil
	}
	if !isCondition {
		if vType == reflect.Slice {
			arr := value.([]interface{})
			if len(arr) == 1 {
				if v := whereSliceValue(fieldOrCondition, []interface{}{"=", arr[0]}); v != nil {
					return v
				}
				return nil
			} else if len(arr) > 0 {
				if s, ok := arr[0].(string); !ok {
					return nil
				} else {
					if v := whereSliceValue(fieldOrCondition, []interface{}{s, arr[1]}); v != nil {
						return v
					}
				}
			}
			return nil
		} else if vType == reflect.Map {
			v := whereMap(t.New(value).MapStringInterface())
			if v != nil {
				return map[string]interface{}{fieldOrCondition: v }
			}
			return nil
		}
		return map[string]interface{}{fieldOrCondition: []interface{}{"=", value}}
	}
	// condition
	arr, ok := value.([]interface{})
	if ok {
		return whereCondition(condition, arr)
	}
	m, ok := value.(map[string]interface{})
	if ok {
		if v := whereMap(m); v != nil {
			return map[string]interface{}{condition: v}
		}
	}
	return nil
}

func whereP3(field string, compare string, value interface{}) map[string]interface{} {
	if field == "" {
		return nil
	}
	return map[string]interface{}{field: []interface{}{ compare, value}}
}

func whereMap(express map[string]interface{}) []interface{} {
	var results []interface{}
	for fd, v := range express {
		if strings.ToLower(fd) == "$condition" {
			continue
		}
		vType := reflect.TypeOf(v).Kind()
		switch fd {
		case "$string":
			results = append(results, map[string]interface{}{fd: v })
			continue
		case "$or", "$and":
			if vType != reflect.Slice && vType != reflect.Map {
				continue
			}
			if vType == reflect.Slice {
				arr := t.New(v).SliceInterface()
				if vItem := whereCondition(fd, arr); vItem != nil {
					results = append(results, map[string]interface{}{fd: vItem })
				}
				continue
			}
			if vType == reflect.Map {
				m := t.New(v).MapStringInterface()
				if mItems := whereMap(m); mItems != nil {
					results = append(results, map[string]interface{}{fd: mItems})
				}
			}
			break
		}
		switch vType {
		case reflect.Slice:
			arr := t.New(v).SliceInterface()
			if arr == nil {
				continue
			}
			if v := whereSliceValue(fd, arr); v != nil {
				results = append(results, v)
			}
			break
		case reflect.Map:
			m1 := t.New(v).MapStringInterface()
			if m1 != nil {
				if v1 := whereMap(m1); v1 != nil {
					results = append(results, map[string]interface{}{
						fd: v1,
					})
				}
			}
			continue
		case reflect.Struct:
			if o, ok := v.(IToString); ok {
				s := o.String()
				results = append(results, map[string]interface{}{ "$string": s })
			}
			break
		case reflect.Func:
			break
		default:
			results = append(results, map[string]interface{}{fd:[]interface{}{ "=", v }})
			break
		}
	}
	if len(results) == 0 {
		return nil
	}
	return results
}

func CheckWhere(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return nil
	}
	argLen := len(args)
	switch argLen {
	case 3:
		fd, ok1 := args[0].(string)
		cp, ok2 := args[1].(string)
		if ok1 && ok2 {
			v := whereP3(fd, cp, args[2])
			if v != nil {
				return []interface{}{v}
			}
		}
		break
	case 2:
		if s, ok := args[0].(string); ok {
			v := whereP2(s, args[1])
			if v == nil {
				return nil
			}
			switch where := v.(type) {
			case []interface{}:
				return where
			case map[string]interface{}:
				return []interface{}{where}
			}
		}
		break
	case 1:
		v := whereP1(args[0])
		if v == nil {
			return nil
		}
		switch where := v.(type) {
		case []interface{}:
			return where
		case map[string]interface{}:
			return []interface{}{where}
		}
		break
	}
	return nil
}
