package sqlw

import (
	"fmt"
	"github.com/gohouse/t"
	"reflect"
	"strings"
)

// 生成字符串条件参数
func buildStrParam(placeHolder string, param string) (string, []interface{}) {
	return param, nil
}

func buildWhereIn(placeHolder string, logic string, field string, values []interface{}) (string, []interface{}) {
	var pms []string
	var vars []interface{}
	operation := "in"
	if logic != "" {
		operation = logic + " " + operation
	}
	for _, v := range values {
		pms = append(pms, placeHolder)
		vars = append(vars, v)
	}
	if len(pms) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%s %s(%s)", field, operation, strings.Join(pms, ",")), vars
}

func buildWhereBetween(placeHolder string, logic string, field string, values []interface{}) (string, []interface{}) {
	varLen := len(values)
	if varLen == 0 {
		return "", nil
	}
	operation := "between"
	if logic != "" {
		operation = logic + " " + operation
	}
	if varLen == 1 {
		return fmt.Sprintf("%s %s (%s and %s)", field, operation, placeHolder, placeHolder), []interface{}{values[0], values[0]}
	}
	return fmt.Sprintf("%s %s (%s and %s)", field, operation, placeHolder, placeHolder), values
}

func buildVarParam(placeHolder string, field string, express []interface{}) (string, []interface{})  {
	if len(express) != 2 {
		return "", nil
	}
	cp, ok := express[0].(string)
	if !ok {
		return "", nil
	}
	varType := reflect.TypeOf(express[1]).Kind()
	var value interface{}
	relation := strings.Trim(strings.ToLower(cp), " ")

	if relation == "in" || relation == "not in" || relation == "between" || relation == "not between" {
		if varType != reflect.Slice {
			return "", nil
		}
		value = t.New(express[1]).SliceInterface()
	} else {
		value = express[1]
	}
	var szWhere = ""
	var vars []interface{}
	switch relation {
	case "in":
		szWhere, vars = buildWhereIn(placeHolder, "", field, value.([]interface{}))
		break
	case "not in":
		szWhere, vars = buildWhereIn(placeHolder, "not", field, value.([]interface{}))
		break
	case "between":
		szWhere, vars = buildWhereBetween(placeHolder, "", field, value.([]interface{}))
		break
	case "not between":
		szWhere, vars = buildWhereBetween(placeHolder, "not", field, value.([]interface{}))
		break
	default:
		szWhere = fmt.Sprintf("%s %s %s", field, cp, placeHolder)
		vars = append(vars, value)
	}
	if szWhere == "" {
		return "", nil
	}
	return szWhere, vars
}

func buildCondition(placeHolder string, condition string, wheres []interface{}) (string, []interface{}) {
	var resultVars []interface{}
	var results []string

	var pushWhereItem = func(where string, vars ...interface{}) {
		if where != "" {
			results = append(results, where)
		}
		if vars != nil {
			resultVars = append(resultVars, vars...)
		}
	}
	for _, item := range wheres {
		keyVars, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		for fdOrCd, value := range keyVars {
			switch fdOrCd {
			case "$string":
				if s, ok := value.(string); ok {
					strWhere, vars := buildStrParam(placeHolder, s)
					pushWhereItem(strWhere, vars...)
				}
				break
			case "$and", "$or":
				// 条件表达式
				vType := reflect.TypeOf(value).Kind()
				if vType != reflect.Slice {
					continue
				}
				params := t.New(value).SliceInterface()
				cdWhere, vars := buildCondition(placeHolder, fdOrCd, params)
				if cdWhere != "" {
					pushWhereItem("(" + cdWhere + ")", vars...)
				}
				break
			default:
				varType := reflect.TypeOf(value).Kind()
				var params []interface{}
				if varType == reflect.Slice {
					params = t.New(value).SliceInterface()
				} else if varType == reflect.Map || varType == reflect.Struct {
					continue
				} else {
					params = []interface{}{"=", value}
				}
				varWhere, vars := buildVarParam(placeHolder, fdOrCd, params)
				pushWhereItem(varWhere, vars...)
				break
			}
		}
	}
	relation := " and "
	if condition == "$or" {
		relation = " or "
	}
	return strings.Join(results, relation), resultVars
}

// 生成sql条件表达式
// @param placeHolder string 参数占位符
// @wheres 条件参数
// return 参数1： where 条件表达式字符串，参数2参数列表
func BuildWhereSql(placeHolder string, wheres []interface{}) (string, []interface{}) {
	return buildCondition(placeHolder, "$and", wheres)
}
