package sqlw

import (
	"encoding/json"
	"testing"
)

var (
	oldWhere = [][]interface{}{
		{"name", "like", "sss%"},
		{"status = 1"},
		{"sex", "in", []int{0, 1, 2}},
		{"age", ">", 3},
		{"$or", []interface{}{
				[]interface{}{"standard", "like", "test%"},
				[]interface{}{"standard", "=", "demo" },
			},
		},
		{ "$and", []interface{}{
				[]interface{}{"model", ">", 10},
				[]interface{}{"model", "<", 100},
			},
		},
		{ "cr_date", "between", []string{ "2019-09-01", "2019-09-30" }},
		{ "$or", map[string]interface{}{
				"price": []interface{}{">=", 30 },
				"saleCount": 12,
			},
		},
	}

	mapWhere = [][]interface{}{
		{  "$or",
			map[string]interface{}{
				"price": []interface{}{">=", 30 },
				"saleCount": 12,
				"$and": map[string]interface{}{
					"integral": []interface{}{"<", 100},
					"sex": []interface{}{"in", []int{1,2}},
				},
			},
		},
	}

	where3 = []interface{}{
		[]interface{}{"age", ">", 30 },
		[]interface{}{"weight", "between", []int{45, 80} },
		[]interface{}{"name", "like", "张%"},
		[]interface{}{"sex", "in", []int{1,2}},
		[]interface{}{"$or",
			[]interface{}{
				[]interface{}{"audit", 1},
				[]interface{}{"status", ">", 2 },
			},
		},
	}

	oldWhereBuff []interface{}
	mapWhereBuff []interface{}
	where3Buff []interface{}
)

func TestCheckWhere(t *testing.T) {
	oldWhereBuff = CheckWhere(oldWhere)
	bytes, err := json.MarshalIndent(oldWhereBuff, "", " ")
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bytes))

	mapWhereBuff = CheckWhere(mapWhere)
	bytes2, err2 := json.MarshalIndent(mapWhereBuff, "", " ")
	if err2 != nil {
		t.Error(err2)
	}
	t.Log(string(bytes2))

	where3Buff = CheckWhere(where3)
	bytes2, err2 = json.MarshalIndent(where3Buff, "", " ")
	if err2 != nil {
		t.Error(err2)
	}
	t.Log(string(bytes2))
}

func TestBuildWhereSql(t *testing.T) {
	where, vars := BuildWhereSql("?", oldWhereBuff)
	t.Logf("where:%s, vars:%v\n", where, vars)

	where, vars = BuildWhereSql("?", mapWhereBuff)
	t.Logf("where:%s, vars:%v\n", where, vars)

	where, vars = BuildWhereSql("?", where3Buff)
	t.Logf("where:%s, vars:%v\n", where, vars)
}

