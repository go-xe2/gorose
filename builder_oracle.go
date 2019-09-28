package gorose

import (
	"errors"
	"fmt"
	"github.com/gohouse/gocar/structEngin"
	"github.com/gohouse/t"
	"github.com/x-goe/gorose/sqlw"

	//"log"
	"reflect"
	"strconv"
	"strings"
)

const (
	DriverOracle = "oci8" // 默认驱动
)

var (
	operatorOracle = []string{"=", ">", "<", "!=", "<>", ">=", "<=", "like", "not like",
		"intersect", "minus", "union", "||", "in", "not in", "between", "not between"}
)

type BuilderOracle struct {
	BuilderDefault
}

var builderOracle *BuilderOracle

func NewBuilderOracle(o IOrm) *BuilderOracle {
	onceBuilderDefault.Do(func() {
		builderOracle = new(BuilderOracle)
		builderOracle.operator = operatorOracle
	})
	builderOracle.IOrm = o
	// 每次使用的时候, 重置为0, 方便pg的占位符使用
	builderOracle.placeholder = 0

	return builderOracle
}

func init() {
	var builder = &BuilderOracle{}
	NewBuilderDriver().Register(DriverOracle, builder)
}

// SetDriver 设置驱动, 方便获取占位符使用
func (b *BuilderOracle) SetDriver(dr string) *BuilderOracle {
	b.driver = dr
	return b
}

// GetPlaceholder 获取占位符
func (b *BuilderOracle) GetPlaceholder() (phstr string) {
	withLockContext(func() {
		ph := b.placeholder + 1
		phstr = fmt.Sprintf(":%v", ph)
		b.placeholder = ph
	})
	return
}

func (b *BuilderOracle) BuildQueryOra() (sqlStr string, args []interface{}, err error) {
	//b.IOrm = o
	join, err := b.BuildJoin()
	if err != nil {
		b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
		return
	}
	where, err := b.BuildWhere()
	if err != nil {
		b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
		return
	}

	if len(b.GetJoin()) > 0 {
		b.GetFields()
	}

	// 默认情况
	fieldsStr := b.BuildFields()
	tableName := b.BuildTable()
	sqlStr = fmt.Sprintf("SELECT %s%s FROM %s%s%s%s%s%s%s", b.BuildDistinct(), fieldsStr,
		tableName, join, where, b.BuildLimit(), b.BuildGroup(), b.BuildHaving(), b.BuildOrder())

	// 批量取数据需嵌套写法
	if b.GetLimit() > 0 {
		aliasNameA := "tabA"
		aliasNameB := "tabB"
		page := b.GetOffset()/b.GetLimit() + 1
		startRow := (page-1)*b.GetLimit() + 1
		endRow := page*b.GetLimit() + 1

		if fieldsStr == "*" {
			fieldsStr = b.GetTable() + ".*, rownum r"
		} else {
			if b.GetGroup() == "" {
				fieldsStr = fieldsStr + ", rownum r"
			}
		}

		// 没有group by需要1层嵌套， 有group by需要2层嵌套
		// 如果考虑orderby优化，还需要一层嵌套。目前未考虑
		if b.GetGroup() == "" {
			sqlStr = fmt.Sprintf("SELECT %s%s FROM %s%s%s%s%s", b.BuildDistinct(), fieldsStr,
				tableName, join, where, b.BuildLimit(), b.BuildOrder())

			sqlStr = fmt.Sprintf("select * from (%s) %s where %s.r>=%s",
				sqlStr, aliasNameA, aliasNameA, strconv.Itoa(startRow))
		} else {
			sqlStr = fmt.Sprintf("SELECT %s%s FROM %s%s%s%s%s%s", b.BuildDistinct(), fieldsStr,
				tableName, join, where, b.BuildGroup(), b.BuildHaving(), b.BuildOrder())

			sqlStr = fmt.Sprintf(
				"select * from (select %s, rownum r from (%s) %s where rownum<%s ) %s where %s.r>=%s",
				aliasNameA+".*", sqlStr, aliasNameA, strconv.Itoa(endRow), aliasNameB, aliasNameB,
				strconv.Itoa(startRow))
		}
	}

	//args = b.bindParams
	args = b.IOrm.GetBindValues()
	return
}

func (b *BuilderOracle) BuildExecuteOra(operType string) (sqlStr string, args []interface{}, err error) {
	// insert : {"name":"fizz, "website":"fizzday.net"} or {{"name":"fizz2", "website":"www.fizzday.net"}, {"name":"fizz", "website":"fizzday.net"}}}
	// update : {"name":"fizz", "website":"fizzday.net"}
	// delete : ...
	//b.IOrm = o
	var update, insertkey, insertval string
	if operType != "delete" {
		if b.IOrm.GetData() == nil {
			err = errors.New("insert,update请传入数据操作")
			b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
			return
		}
		update, insertkey, insertval = b.BuildData(operType)
	}

	where, err := b.BuildWhere()
	if err != nil {
		b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
		return
	}

	switch operType {
	case "insert":
		sqlStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", b.BuildTable(), insertkey, insertval)
	case "update":
		if where == "" && b.IOrm.GetForce() == false {
			err = errors.New("出于安全考虑, update时where条件不能为空, 如果真的不需要where条件, 请使用Force()(如: db.xxx.Force().Update())")
			b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
			return
		}
		sqlStr = fmt.Sprintf("UPDATE %s SET %s%s", b.BuildTable(), update, where)
	case "delete":
		if where == "" && b.IOrm.GetForce() == false {
			err = errors.New("出于安全考虑, delete时where条件不能为空, 如果真的不需要where条件, 请使用Force()(如: db.xxx.Force().Delete())")
			b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
			return
		}
		sqlStr = fmt.Sprintf("DELETE FROM %s%s", b.BuildTable(), where)
	}

	args = b.IOrm.GetBindValues()
	return
}

func (b *BuilderOracle) BuildData(operType string) (string, string, string) {
	data := b.IOrm.GetData()
	ref := reflect.Indirect(reflect.ValueOf(data))

	switch ref.Kind() {
	case reflect.Struct:
		return b.parseData(operType, structEngin.New().SetExtraCols(b.IOrm.GetExtraCols()).StructContent2Map(data))
	case reflect.Map:
		var tmp = []map[string]interface{}{t.New(data).MapStringInterface()}
		return b.parseData(operType, tmp)
	case reflect.Slice:
		switch ref.Type().Elem().Kind() {
		case reflect.Struct:
			return b.parseData(operType, structEngin.New().SetExtraCols(b.IOrm.GetExtraCols()).StructContent2Map(data))
		case reflect.Map:
			return b.parseData(operType, t.New(data).SliceMapStringInterface())
		}
	}
	return "", "", ""
}

func (b *BuilderOracle) BuildData2(operType string) (string, string, string) {
	return b.BuilderDefault.BuildData2(operType)
}

func (b *BuilderOracle) parseData(operType string, data []map[string]interface{}) (string, string, string) {
	// insert
	var dataFields []string
	var dataValues []string
	// update or delete
	var dataObj []string

	for key, _ := range data[0] {
		if inArray(key, dataFields) == false {
			dataFields = append(dataFields, key)
		}
	}
	for _, item := range data {
		// 定义1条数据的存储
		var dataValuesSub []string
		for _, key := range dataFields {
			if item[key] == nil {
				// 放入占位符
				dataValuesSub = append(dataValuesSub, b.GetPlaceholder())
				// 保存真正的值为null
				b.IOrm.SetBindValues("null")
			} else {
				// 放入占位符
				dataValuesSub = append(dataValuesSub, b.GetPlaceholder())
				// 保存真正的值
				b.IOrm.SetBindValues(item[key])
			}
			// update
			dataObj = append(dataObj, fmt.Sprintf("%s=%s", key, b.GetPlaceholder()))
		}
		dataValues = append(dataValues, "("+strings.Join(dataValuesSub, ",")+")")
	}
	return strings.Join(dataObj, ","), strings.Join(dataFields, ","), strings.Join(dataValues, ",")
}

func (b *BuilderOracle) BuildJoin() (s string, err error) {
	return b.BuilderDefault.BuildJoin()
}

func (b *BuilderOracle) BuildWhere() (where string, err error) {
	var beforeParseWhere = b.IOrm.GetWhere()
	where, err = b.parseWhere(b.IOrm)
	b.IOrm.SetWhere(beforeParseWhere)
	return If(where == "", "", " WHERE "+where).(string), err
}

func (b *BuilderOracle) BuildDistinct() (dis string) {
	return b.BuilderDefault.BuildDistinct()
}

func (b *BuilderOracle) BuildFields() string {
	return b.BuilderDefault.BuildFields()
}

func (b *BuilderOracle) BuildTable() string {
	return b.BuilderDefault.BuildTable()
}

func (b *BuilderOracle) BuildGroup() string {
	return b.BuilderDefault.BuildGroup()
}

func (b *BuilderOracle) BuildHaving() string {
	return b.BuilderDefault.BuildHaving()
}

func (b *BuilderOracle) BuildOrder() string {
	return b.BuilderDefault.BuildOrder()
}

func (b *BuilderOracle) BuildLimit() string {
	if b.IOrm.GetUnion() != nil {
		return ""
	}

	if b.GetLimit() == 0 {
		return ""
	}

	page := b.GetOffset()/b.GetLimit() + 1
	endRow := page*b.GetLimit() + 1

	var limitStr string
	if len(b.IOrm.GetWhere()) > 0 {
		limitStr = fmt.Sprintf(" and rownum < %d", endRow)
	} else {
		limitStr = fmt.Sprintf(" where rownum < %d", endRow)
	}
	return If(b.IOrm.GetLimit() == 0, "", limitStr).(string)
}

func (b *BuilderOracle) BuildOffset() string {
	return ""
}

func (b *BuilderOracle) parseWhere(ormApi IOrm) (string, error) {
	// 取出所有where
	wheres := ormApi.GetWhere()
	where, vars := sqlw.BuildWhereSql(b.GetPlaceholder(), wheres)
	if where == "" {
		return "", nil
	}
	b.SetBindValues(vars...)
	return where, nil
}

func (b *BuilderOracle) parseParams(args []interface{}, ormApi IOrm) (s string, err error) {
	paramsLength := len(args)
	argsReal := args

	// 存储当前所有数据的数组
	var paramsToArr []string

	switch paramsLength {
	case 3: // 常规3个参数:  {"id",">",1}
		//if !inArray(argsReal[1], b.GetRegex()) {
		if !inArray(argsReal[1], b.GetOperator()) {
			err = errors.New("where parameter is wrong")
			b.IOrm.GetISession().GetIEngin().GetLogger().Error(err.Error())
			return
		}

		paramsToArr = append(paramsToArr, argsReal[0].(string))
		paramsToArr = append(paramsToArr, argsReal[1].(string))

		switch argsReal[1] {
		case "like", "not like":
			paramsToArr = append(paramsToArr, b.GetPlaceholder())
			b.IOrm.SetBindValues(argsReal[2])
		case "in", "not in":
			var tmp []string
			var ar2 = t.New(argsReal[2]).Slice()
			for _, item := range ar2 {
				tmp = append(tmp, b.GetPlaceholder())
				b.IOrm.SetBindValues(t.New(item).Interface())
			}
			paramsToArr = append(paramsToArr, "("+strings.Join(tmp, ",")+")")
		case "between", "not between":
			var ar2 = t.New(argsReal[2]).Slice()
			paramsToArr = append(paramsToArr, b.GetPlaceholder()+" and "+b.GetPlaceholder())
			b.IOrm.SetBindValues(ar2[0].Interface())
			b.IOrm.SetBindValues(ar2[1].Interface())
		default:
			paramsToArr = append(paramsToArr, b.GetPlaceholder())
			b.IOrm.SetBindValues(argsReal[2])
		}
	case 2:
		paramsToArr = append(paramsToArr, argsReal[0].(string))
		paramsToArr = append(paramsToArr, "=")
		paramsToArr = append(paramsToArr, b.GetPlaceholder())
		b.IOrm.SetBindValues(argsReal[1])
	}

	return strings.Join(paramsToArr, " "), nil
}

func (b *BuilderOracle) GetOperator() []string {
	return b.BuilderDefault.GetOperator()
}

// 实现接口
// BuildQuery : build query sql string
func (b *BuilderOracle) BuildQuery(o IOrm) (sqlStr string, args []interface{}, err error) {
	return NewBuilderOracle(o).SetDriver(DriverOracle).BuildQueryOra()
}

// BuildExecut : build execute sql string
func (b *BuilderOracle) BuildExecute(o IOrm, operType string) (sqlStr string, args []interface{}, err error) {
	return NewBuilderOracle(o).SetDriver(DriverOracle).BuildExecuteOra(operType)
}
