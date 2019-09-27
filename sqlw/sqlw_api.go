package sqlw

import (
	"fmt"
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
			m, ok := arg.(map[string]interface{})
			if !ok {
				continue
			}
			results = append(results, m)
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
		arr, ok := express.([]interface{})
		if ok {
			return whereCondition("$and", arr)
		}
		arr1, ok := express.([][]interface{})
		if !ok {
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
	fmt.Printf("condition:%v,value:%v\n", fieldOrCondition, value)

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
		if fd == "$string" {
			results = append(results, map[string]interface{}{fd: v })
			continue
		}
		vType := reflect.TypeOf(v).Kind()
		switch vType {
		case reflect.Slice:
			arr, ok := v.([]interface{})
			if !ok {
				continue
			}
			arrLen := len(arr)
			if arrLen == 1 {
				if v := whereP1(arr[0]); v != nil {
					results = append(results, v)
				}
			} else if arrLen > 1 {
				if v := whereSliceValue(fd, arr); v != nil {
					results = append(results, v)
				}
			}
			break
		case reflect.Map:
			m1, ok := v.(map[string]interface{})
			if ok {
				if v1 := whereMap(m1); v1 != nil {
					results = append(results, map[string]interface{}{
						fd: v,
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
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32,
				reflect.Float64, reflect.String:
					fmt.Printf("real value:%v\n", v)
			results = append(results, map[string]interface{}{fd:[]interface{}{ "=", v }})
			break
		case reflect.Uintptr:
			results = append(results, map[string]interface{}{fd:[]interface{}{ "=", *v.(*uint) }})
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
