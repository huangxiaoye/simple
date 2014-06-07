package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	. "simple/simple"
)

var (
	RuleName = &StringRule{"name", 4, 20}
	RuleAge  = &IntRule{"age", 10, 60}
)

func authorize(pattern string, request *http.Request, params Value) bool {
	return params.String("name") == "user1"
}

func test(params Value) Value {
	fmt.Println("in:", params)

	//适用男规则验证输入参数
	if !RuleCheck(params, RuleName, RuleAge) {
		return NewValue().Error("invalid params")
	}

	//返回数据
	params["__time__"] = time.Now()
	return params
}

func main() {
	s := NewServer(":8080")

	s.StaticDir = "."       // 静态文件目录。
	s.Pattern = "/v2/"      // REST 路径前缀匹配模式。
	s.Authorize = authorize // 请求验证函数。

	s.HandleFunc("!/test/{name}/{age}", GET, test) // http://localhost:8080/v2/test/user1/23

	if err := s.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
