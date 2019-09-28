package xorm

import (
	"testing"
)

type O = map[string]interface{}
type A = []interface{}

func DB() IOrm {
	return initDB().NewOrm()
}
func TestNewOrm(t *testing.T) {
	orm := DB()
	orm.Hello()
}
func TestOrm_AddFields(t *testing.T) {
	orm := DB()
	//var u = Users{}
	var fieldStmt = orm.Table("users").Fields("a").Where("m", 55)
	a, b, err := fieldStmt.AddFields("b").Where("d", 1).BuildSql()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(a, b)

	fieldStmt.Reset()
	d, e, err := fieldStmt.Fields("a").AddFields("c").Where("d", 2).BuildSql()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(d, e)
}

func TestOrm_BuildSql(t *testing.T) {
	var u = Users{
		Name: "gorose2",
		Age:  19,
	}

	//aff, err := db.Force().Data(&u)
	a, b, err := DB().Table(&u).Where("age", ">", 1).Data(&u).BuildSql("update")
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(a, b)
}

func TestOrm_BuildSql_where(t *testing.T) {
	var u = Users{
		Name: "gorose2",
		Age:  19,
	}

	var db = DB()
	a, b, err := db.Table(&u).Where([][]interface{}{{"age", "<>", 1}, {"age", "in", []int{0,1,2}}}).Limit(10).BuildSql()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(a, b)

	db.Reset()
	a, b, err = db.Table(&u).Where(A{
			O{"age": A{"<>", 1}},
			O{"age": A{"in", []int{0,1,2}}},
			O{
				"$or": A{
					O{"name": "张三"},
					O{"name": A{"like", "李%"}},
				},
			},
		}).Limit(10).BuildSql()
	if err != nil {
		t.Error(err.Error())
	}
	t.Log(a, b)
}
