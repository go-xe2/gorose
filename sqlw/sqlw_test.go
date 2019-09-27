package sqlw

import (
	"encoding/json"
	"testing"
)

func TestCheckWhere(t *testing.T) {
	//oldWhere := [][]interface{}{
	//	{"name", "like", "sss%"},
	//	{"status = 1"},
	//	{"sex", "in", []int{0, 1, 2}},
	//	{"age", ">", 3},
	//	{"$or", []interface{}{
	//			[]interface{}{"standard", "like", "test%"},
	//			[]interface{}{"standard", "=", "demo" },
	//		},
	//	},
	//	{ "$and", []interface{}{
	//			[]interface{}{"model", ">", 10},
	//			[]interface{}{"model", "<", 100},
	//		},
	//	},
	//	{ "cr_date", "between", []string{ "2019-09-01", "2019-09-30" }},
	//	{ "$or", map[string]interface{}{
	//			"price": []interface{}{">=", 30 },
	//			"saleCount": 12,
	//		},
	//	},
	//}
	//
	//arr1 := CheckWhere(oldWhere)
	//bytes, err := json.MarshalIndent(arr1, "", " ")
	//if err != nil {
	//	t.Error(err)
	//}
	//t.Log(string(bytes))

	mapWhere := [][]interface{}{
		{ "$or", map[string]interface{}{
				"price": []interface{}{">=", 30 },
				"saleCount": 12,
				"$and": map[string]interface{}{
					"integral": []interface{}{"<", 100},
					"sex": []interface{}{"in", []int{1,2}},
				},
			},
		},
	}

	arr2 := CheckWhere(mapWhere)
	bytes2, err2 := json.MarshalIndent(arr2, "", " ")
	if err2 != nil {
		t.Error(err2)
	}
	t.Log(string(bytes2))
}
