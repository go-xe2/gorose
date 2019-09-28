# xorm orm

## 安装使用
- go.mod
```bash
require github.com/go-xe2/xorm
```
> 使用 `import "github.com/go-xe2/xorm"`导入库 

- go get  
```bash
go get -u github.com/go-xe2/xorm
```

## 该ORM由gohouse/gorose改版升级，原gorose文档请参数:
[2.x doc](https://www.kancloud.cn/fizz/gorose-2/1135835)  

## api预览
```go
db.Table().Fields().Distinct().Where().GroupBy().Having().OrderBy().Limit().Offset().Select()
db.Table().Data().Insert()
db.Table().Data().Where().Update()
db.Table().Where().Delete()
```

## 使用示例：
```go
package main
import (
	"fmt"
	"github.com/go-xe2/xorm"
	_ "github.com/mattn/go-sqlite3"
)
var err error
var engin *xorm.Engin
func init() {
    // Global initialization and reuse of databases
    // The engin here needs to be saved globally, using either global variables or singletons
    // Configuration & xorm. Config {} is a single database configuration
    // If you configure a read-write separation cluster, use & xorm. ConfigCluster {}
	engin, err = xorm.Open(&xorm.Config{Driver: "sqlite3", Dsn: "./db.sqlite"})
    // mysql demo, remeber import mysql driver of github.com/go-sql-driver/mysql
	// engin, err = xorm.Open(&xorm.Config{Driver: "mysql", Dsn: "root:root@tcp(localhost:3306)/test?charset=utf8&parseTime=true"})
}
func DB() xorm.IOrm {
	return engin.NewOrm()
}
func main() {
    // Native SQL, return results directly 
    res,err := DB().Query("select * from users where uid>? limit 2", 1)
    fmt.Println(res)
    affected_rows,err := DB().Execute("delete from users where uid=?", 1)
    fmt.Println(affected_rows, err)

    // orm chan operation, fetch one row
    res, err := DB().Table("users").First()
    // res's type is map[string]interface{}
    fmt.Println(res)
    
    // rm chan operation, fetch more rows
    res2, _ := DB().Table("users").Get()
    // res2's type is []map[string]interface{}
    fmt.Println(res2)
    
    where := []interface{}{
         []interface{}{"age", ">", 30 },
         []interface{}{"weight", "between", []int{45, 80} },
         []interface{}{"name", "like", "张%"},
         []interface{}{"sex", "in", []int{1,2}},
         []interface{}{"$or",
             []interface{}{
                 []interface{}{"audit", 1},
                 []interface{}{"status", ">", 2 },
             },
    }
    res3, _ := DB().Table("test").Where(where).Get()
    fmt.Println(res3)
}
```

## 配置示例:
```go
var configSimple = &xorm.Config{
	Driver: "sqlite3", 
	Dsn: "./db.sqlite",
}
```
配置:
```go
var config = &xorm.ConfigCluster{
	Master:       []&xorm.Config{}{configSimple}
    Slave:        []&xorm.Config{}{configSimple}
    Prefix:       "pre_",
    Driver:       "sqlite3",
}
```
初始化:
```go
var engin *xorm.Engin
engin, err := Open(config)

if err != nil {
    panic(err.Error())
}
```

## Native SQL operation (add, delete, check), session usage
创建用户表: `users`
```sql
DROP TABLE IF EXISTS "users";
CREATE TABLE "users" (
	 "uid" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	 "name" TEXT NOT NULL,
	 "age" integer NOT NULL
);

INSERT INTO "users" VALUES (1, 'xe222', 18);
INSERT INTO "users" VALUES (2, 'xe2', 18);
INSERT INTO "users" VALUES (3, 'fizzday', 18);
```
定义实体:
```go
type Users struct {
	Uid  int    `orm:"uid"`
	Name string `orm:"name"`
	Age  int    `orm:"age"`
}
// 实体对应数据库表名
func (u *Users) TableName() string {
	return "users"
}
```
原生sql查询:
```go
// Here is the structure object to be bound
// If you don't define a structure, you can use map, map example directly
// var u = xorm.Data{}
// var u = xorm.Map{}  Both are possible.
var u Users
session := engin.NewSession()
// Here Bind () is used to store results. If you use NewOrm () initialization, you can use NewOrm (). Table (). Query () directly.
_,err := session.Bind(&u).Query("select * from users where uid=? limit 2", 1)
fmt.Println(err)
fmt.Println(u)
fmt.Println(session.LastSql())
```
原生sql删改查:
```go
session.Execute("insert into users(name,age) values(?,?)(?,?)", "xorm",18,"fizzday",19)
session.Execute("update users set name=? where uid=?","xorm",1)
session.Execute("delete from users where uid=?", 1)
```

- 1. 基本使用方法  

```go
var u Users
db := engin.NewOrm()
err := db.Table(&u).Fields("name").AddFields("uid","age").Distinct().Where("uid",">",0).OrWhere("age",18).
	Group("age").Having("age>1").OrderBy("uid desc").Limit(10).Offset(1).Select()
```

- 2. 使用map
```go
type user xorm.Map
// Or the following type definitions can be parsed properly
type user2 map[string]interface{}
type users3 []user
type users4 []map[string]string
type users5 []xorm.Map
type users6 []xorm.Data
```
- 3、定义实体
```go
db.Table(&user).Select()
db.Table(&users4).Limit(5).Select()
```

- 4. 删除改查 
```go
db.Table(&user2).Limit(10.Select()
db.Table(&user2).Where("uid", 1).Data(xorm.Data{"name","xorm"}).Update()
db.Table(&user2).Data(xorm.Data{"name","xe2"}).Insert()
db.Table(&user2).Data([]xorm.Data{{"name","xe23"},"name","dddd"}).Insert()
db.Table(&user2).Where("uid", 1).Delete()
```

## 共享连接池
sql
```sql
DROP TABLE IF EXISTS "users";
CREATE TABLE "users" (
	 "uid" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	 "name" TEXT NOT NULL,
	 "age" integer NOT NULL
);

INSERT INTO "users" VALUES (1, 'txt', 18);
INSERT INTO "users" VALUES (2, 'ytx2', 18);
INSERT INTO "users" VALUES (3, 'fizzday', 18);
```
Actual Code
```go
package main

import (
	"fmt"
	"github.com/go-xe2/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type Users struct {
    Uid int64 `orm:"uid"`
    Name string `orm:"name"`
    Age int64 `orm:"age"`
    Xxx interface{} `orm:"-"` // This field is ignored in ORM
}

func (u *Users) TableName() string {
	return "users"
}

var err error
var engin *xorm.Engin

func init() {
    // Global initialization and reuse of databases
    // The engin here needs to be saved globally, using either global variables or singletons
    // Configuration & xorm. Config {} is a single database configuration
    // If you configure a read-write separation cluster, use & xorm. ConfigCluster {}
	engin, err = xorm.Open(&xorm.Config{Driver: "sqlite3", Dsn: "./db.sqlite"})
}
func DB() xorm.IOrm {
	return engin.NewOrm()
}
func main() {
	// A variable DB is defined here to reuse the DB object, and you can use db. LastSql () to get the SQL that was executed last.
	// If you don't reuse db, but use DB () directly, you create a new ORM object, which is brand new every time.
	// So reusing DB must be within the current session cycle
	db := DB()
	
	// fetch a row
	var u Users
	// bind result to user{}
	err = db.Table(&u).Fields("uid,name,age").Where("age",">",0).OrderBy("uid desc").Select()
	if err!=nil {
		fmt.Println(err)
	}
	fmt.Println(u, u.Name)
	fmt.Println(db.LastSql())
	
	// fetch multi rows
	// bind result to []Users, db and context condition parameters are reused here
	// If you don't want to reuse, you can use DB() to open a new session, or db.Reset()
	// db.Reset() only removes contextual parameter interference, does not change links, DB() will change links.
	var u2 []Users
	err = db.Limit(10).Offset(1).Select()
	fmt.Println(u2)
	
	// count
	var count int64
	// Here reset clears the parameter interference of the upper query and can count all the data. If it is not clear, the condition is the condition of the upper query.
	// At the same time, DB () can be called new, without interference.
	count,err = db.Reset().Count()
	// or
	count, err = DB().Table(&u).Count()
	fmt.Println(count, err)
}
```

## 基本使用

- Chunk Data Fragmentation, Mass Data Batch Processing (Cumulative Processing)  

   ` When a large amount of data needs to be manipulated, the chunk method can be used if it is unreasonable to take it out at one time and then operate it again.  
        The first parameter of chunk is the amount of data specified for a single operation. According to the volume of business, 100 or 1000 can be selected.  
        The second parameter of chunk is a callback method for writing normal data processing logic  
        The goal is to process large amounts of data senselessly  
        The principle of implementation is that each operation automatically records the current operation position, and the next time the data is retrieved again, the data is retrieved from the current position.
        `
	```go
	User := db.Table("users")
	User.Fields("id, name").Where("id",">",2).Chunk(2, func(data []map[string]interface{}) {
	    // for _,item := range data {
	    // 	   fmt.Println(item)
	    // }
	    fmt.Println(data)
	})

	// print result:  
	// map[id:3 name:gorose]
	// map[id:4 name:fizzday]
	// map[id:5 name:fizz3]
	// map[id:6 name:gohouse]
	[map[id:3 name:ytx22] map[name:ssss id:4]]
	[map[id:5 name:tttt] map[id:6 name:xe2]]
	```
    
- Loop Data fragmentation, mass data batch processing (from scratch)   

	` Similar to chunk method, the implementation principle is that every operation is to fetch data from the beginning.
	Reason: When we change data, the result of the change may affect the result of our data taking as where condition, so we can use Loop.`
    ```go
	User := db.Table("users")
	User.Fields("id, name").Where("id",">",2).Loop(2, func(data []map[string]interface{}) {
	    // for _,item := range data {
	    // 	   fmt.Println(item)
	    // }
	    // here run update / delete  
	})
	```
 
## 发布版本
- v1.0.x:
    由gohouse/gorose 改版发布

## 更新说明
### 由 github.com/gohouse/gorose/v2 改版升级为github.com/go-xe2/xorm v1.0.0
- db.Table("tablename") 时自动清空where条件及join关联，在同一会话中不需要主动调用db.ResetWhere或db.Reset()方法
- db.Where() 传入nil或宽数组时自动忽略该条件
- db.Where() 可以传入0到3个参数，支持如下形式：
    * 1、简单条件
    ```
    // 1个参数
    db.Where([][]interface{}{"name", "like", "张三%"})
    // sql:
    // where name like '张三%'
    
    db.Where([]interface{}{map[string]interface{}{"name":["like", "张三%"]}）
    // where name like '张三%'
    
    // 2个参数
    db.Where("name", "张三")
    // sql:
    // where name = '张三'
    
    db.Where("name", []interface{}{"like", "张三%"})
    // sql:
    // where name like '张三%'
    
    // 3个参数
    db.Where("name", "like", "张三%")
    // sql:
    // where name like '张三%'
    
    ```
    * 2、复杂条件
    
    ```
    // and 和 or 混合
    // json:
    `
    [
        ["age", ">", 30],
        ["weight", "between", [45,80]],
        ["name", "like", "张%"],
        ["sex", "in", [1,2]],
        ["$or",
            [
                ["audit", 1],
                ["status", 2]
            ]
        ]
    ]
    `
    // golang:
    
    where := []interface{}{
        []interface{}{"age", ">", 30 },
        []interface{}{"weight", "between", []int{45, 80} },
        []interface{}{"name", "like", "张%"},
        []interface{}{"sex", "in", []int{1,2}},
        []interface{}{"$or",
            []interface{}{
                []interface{}{"audit", 1},
                []interface{}{"status", ">", 2 },
            },
    }
    
    db.Where(where)
    
    // sql:
    // where age > 30 and weight between (45 and 80) and name like '张%'
    //      and sex in(1,2) and (audit = 1 or status > 2)
        
    ``` 
