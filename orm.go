package xorm

import (
	"fmt"
	"github.com/go-xe2/xorm/sqlw"
	"strings"
)

type Orm struct {
	ISession
	//IBinder
	*OrmApi
	driver     string
	bindValues []interface{}
}

var _ IOrm = (*Orm)(nil)

func NewOrm(e IEngin) *Orm {
	var orm = new(Orm)
	orm.SetISession(NewSession(e))
	//orm.IBinder = b
	orm.OrmApi = new(OrmApi)
	return orm
}

func (dba *Orm) Hello() {
	fmt.Println("hello gorose orm struct")
}

// ExtraCols 额外的字段
func (dba *Orm) ExtraCols(args ...string) IOrm {
	dba.extraCols = append(dba.extraCols, args...)
	return dba
}

func (dba *Orm) ResetExtraCols() IOrm {
	dba.extraCols = []string{}
	return dba
}

func (dba *Orm) SetBindValues(v ...interface{}) {
	dba.bindValues = append(dba.bindValues, v...)
}

func (dba *Orm) ClearBindValues() {
	dba.bindValues = []interface{}{}
}

func (dba *Orm) GetBindValues() []interface{} {
	return dba.bindValues
}

func (dba *Orm) GetDriver() string {
	return dba.driver
}

func (dba *Orm) SetISession(is ISession) {
	dba.ISession = is
}

func (dba *Orm) GetISession() ISession {
	return dba.ISession
}

// Fields : select fields
func (dba *Orm) Table(tab interface{}) IOrm {
	dba.GetISession().Bind(tab)
	// 重新查询时，应该清除原条件
	dba.ResetWhere()
	dba.join = [][]interface{}{}
	dba.bindValues = []interface{}{}
	//dba.table = dba.GetISession().GetTableName()
	return dba
}

// Fields : select fields
func (dba *Orm) Fields(fields ...string) IOrm {
	dba.fields = fields
	return dba
}

// AddFields : If you already have a query builder instance and you wish to add a column to its existing select clause, you may use the AddFields method:
func (dba *Orm) AddFields(fields ...string) IOrm {
	dba.fields = append(dba.fields, fields...)
	return dba
}

// Distinct : select distinct
func (dba *Orm) Distinct() IOrm {
	dba.distinct = true

	return dba
}

// Data : insert or update data
func (dba *Orm) Data(data interface{}) IOrm {
	dba.data = data
	return dba
}

// Group : select group by
func (dba *Orm) Group(group string) IOrm {
	dba.group = group
	return dba
}

// GroupBy : equals Group()
func (dba *Orm) GroupBy(group string) IOrm {
	return dba.Group(group)
}

// Having : select having
func (dba *Orm) Having(having string) IOrm {
	dba.having = having
	return dba
}

// Order : select order by
func (dba *Orm) Order(order string) IOrm {
	dba.order = order
	return dba
}

// OrderBy : equal order
func (dba *Orm) OrderBy(order string) IOrm {
	return dba.Order(order)
}

// Limit : select limit
func (dba *Orm) Limit(limit int) IOrm {
	dba.limit = limit
	return dba
}

// Offset : select offset
func (dba *Orm) Offset(offset int) IOrm {
	dba.offset = offset
	return dba
}

// Page : select page
func (dba *Orm) Page(page int) IOrm {
	dba.offset = (page - 1) * dba.GetLimit()
	return dba
}


// where 查询条件
// @param args 支持0-3个参数
// 1个时为:slice或map
// 2个时为:参数1：$and|$or|字段名, 参数2: slice|slice|[real value或slice]
// 3个时为:参数1: 字段名， 参数2: >,<,>=,<=, <>, in, not in, between, like 等待关系连接符, 参数3为条件值
func (dba *Orm) Where(args ...interface{}) IOrm {
	where := sqlw.CheckWhere(args...)
	if where != nil {
		dba.where = append(dba.where, where...)
	}
	return dba
}

// 使用or拼装where条件
func (dba *Orm) OrWhere(args ...interface{}) IOrm {
	if len(args) == 0 {
		return dba
	}
	where := sqlw.CheckWhere(args...)
	if where != nil {
		dba.where = append(dba.where, map[string]interface{}{"$or": where})
	}
	return dba
}

func (dba *Orm) WhereNull(arg string) IOrm {
	return dba.Where(arg + " is null")
}

func (dba *Orm) OrWhereNull(arg string) IOrm {
	return dba.OrWhere(arg + " is null")
}

func (dba *Orm) WhereNotNull(arg string) IOrm {
	return dba.Where(arg + " is not null")
}

func (dba *Orm) OrWhereNotNull(arg string) IOrm {
	return dba.OrWhere(arg + " is not null")
}

func (dba *Orm) WhereIn(needle string, hystack []interface{}) IOrm {
	return dba.Where(needle, "in", hystack)
}

func (dba *Orm) OrWhereIn(needle string, hystack []interface{}) IOrm {
	return dba.OrWhere(needle, "in", hystack)
}

func (dba *Orm) WhereNotIn(needle string, hystack []interface{}) IOrm {
	return dba.Where(needle, "not in", hystack)
}

func (dba *Orm) OrWhereNotIn(needle string, hystack []interface{}) IOrm {
	return dba.OrWhere(needle, "not in", hystack)
}

func (dba *Orm) WhereBetween(needle string, hystack []interface{}) IOrm {
	return dba.Where(needle, "between", hystack)
}

func (dba *Orm) OrWhereBetween(needle string, hystack []interface{}) IOrm {
	return dba.OrWhere(needle, "between", hystack)
}

func (dba *Orm) WhereNotBetween(needle string, hystack []interface{}) IOrm {
	return dba.Where(needle, "not between", hystack)
}

func (dba *Orm) OrWhereNotBetween(needle string, hystack []interface{}) IOrm {
	return dba.OrWhere(needle, "not between", hystack)
}

// Join : select join query
func (dba *Orm) Join(joinType string, args ...interface{}) IOrm {
	dba._joinBuilder(joinType, args)
	return dba
}

func (dba *Orm) LeftJoin(args ...interface{}) IOrm {
	dba._joinBuilder("left", args)
	return dba
}
func (dba *Orm) RightJoin(args ...interface{}) IOrm {
	dba._joinBuilder("right", args)
	return dba
}
func (dba *Orm) CrossJoin(args ...interface{}) IOrm {
	dba._joinBuilder("cross", args)
	return dba
}

// _joinBuilder
func (dba *Orm) _joinBuilder(joinType string, args []interface{}) {
	dba.join = append(dba.join, []interface{}{joinType, args})
}

// Reset  orm api and bind values reset to init
func (dba *Orm) Reset() IOrm {
	dba.OrmApi = new(OrmApi)
	dba.ClearBindValues()
	dba.GetISession().SetUnion(nil)
	return dba
}

// ResetWhere
func (dba *Orm) ResetWhere() IOrm {
	dba.where = []interface{}{}
	return dba
}

// ResetUnion
func (dba *Orm) ResetUnion() IOrm {
	dba.union = ""
	dba.GetISession().SetUnion(nil)
	return dba
}

// BuildSql
// operType(select, insert, update, delete)
func (dba *Orm) BuildSql(operType ...string) (a string, b []interface{}, err error) {
	// 解析table
	dba.table, err = dba.GetISession().GetTableName()
	if err != nil {
		dba.GetISession().GetIEngin().GetLogger().Error(err.Error())
		return
	}
	// 解析字段
	// 如果有union操作, 则不需要
	if inArray(dba.GetIBinder().GetBindType(), []interface{}{OBJECT_STRUCT, OBJECT_STRUCT_SLICE}) &&
		dba.GetUnion() == nil {
		dba.fields = getTagName(dba.GetIBinder().GetBindResult(), TAGNAME)
	}
	if len(operType) == 0 || (len(operType) > 0 && strings.ToLower(operType[0]) == "select") {
		// 根据传入的struct, 设置limit, 有效的节约空间
		if dba.union == "" {
			var bindType = dba.GetIBinder().GetBindType()
			if bindType == OBJECT_MAP || bindType == OBJECT_STRUCT {
				dba.Limit(1)
			}
		}
		a, b, err = NewBuilder(dba.GetISession().GetIEngin().GetDriver()).BuildQuery(dba)
		if dba.GetISession().GetTransaction() {
			a = a + dba.GetPessimisticLock()
		}
	} else {
		a, b, err = NewBuilder(dba.GetISession().GetIEngin().GetDriver()).BuildExecute(dba, strings.ToLower(operType[0]))
		// 重置强制获取更新或插入的字段, 防止复用时感染
		dba.ResetExtraCols()
	}
	// 如果是事务, 因为需要复用单一对象, 故参数会产生感染
	// 所以, 在这里做一下数据绑定重置操作
	if dba.GetISession().GetTransaction() {
		dba.Reset()
	}
	return
}

func (s *Orm) Transaction(closers ...func(db IOrm) error) (err error) {
	err = s.ISession.Begin()
	if err != nil {
		s.GetIEngin().GetLogger().Error(err.Error())
		return err
	}

	for _, closer := range closers {
		err = closer(s)
		if err != nil {
			s.GetIEngin().GetLogger().Error(err.Error())
			_ = s.ISession.Rollback()
			return
		}
	}
	return s.ISession.Commit()
}

// SharedLock 共享锁
// select * from xxx lock in share mode
func (dba *Orm) SharedLock() *Orm {
	dba.pessimisticLock = " lock in share mode"
	return dba
}

// LockForUpdate
// select * from xxx for update
func (dba *Orm) LockForUpdate() *Orm {
	dba.pessimisticLock = " for update"
	return dba
}
