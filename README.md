# boot4go-xxljob-executor

golang for xxl-job executor support 

![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

Java and go development coexist in many companies. XXL-job is used as the task scheduling engine in Java. Therefore, go executor (client) also appears, which is relatively simple to use:
# Feature
```	
1. Executor registration
2. Time consuming task cancellation
3. Task registration, like http As convenient as handler
4. Task panic processing
5. Blocking strategy processing
6. Task completion support return execution remarks
7. Task timeout cancellation (unit: seconds, 0 is unlimited)
8. Number of failed retries (in parameter param, the task will handle it by itself at present)
```

# Example
```
package main

import (
	"fmt"
	executor4go "github.com/gohutool/boot4go-xxljob-executor"
	"github.com/gohutool/boot4go-xxljob-executor/examples/task"
	"log"
)

func main() {
	exec := executor4go.NewExecutor(
		executor4go.ServerAddr("http://127.0.0.1/xxl-job-admin"),
		executor4go.AccessToken(""),            //请求令牌(默认为空)
		executor4go.ExecutorIp("127.0.0.1"),    //可自动获取
		executor4go.ExecutorPort("9999"),       //默认9999（非必填）
		executor4go.RegistryKey("golang-jobs"), //执行器名称
		executor4go.SetLogger(&logger{}),       //自定义日志
	)
	executor4go.Init()
	//设置日志查看handler
	executor4go.LogHandler(func(req *executor4go.LogReq) *executor4go.LogRes {
		return &executor4go.LogRes{Code: 200, Msg: "", Content: executor4go.LogResContent{
			FromLineNum: req.FromLineNum,
			ToLineNum:   2,
			LogContent:  "这个是自定义日志handler",
			IsEnd:       true,
		}}
	})
	//注册任务handler
	executor4go.RegTask("task.test", task.Test)
	executor4go.RegTask("task.test2", task.Test2)
	executor4go.RegTask("task.panic", task.Panic)
	log.Fatal(executor4go.Run())
}

//Logger接口实现
type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	log.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}
```
# Sample project
github.com/gohutool/boot4go-xxljob-executor/examples/

# xxl-job-admin configuration
### Add executor
Executor management - > Add Executor. The list of Executors is as follows:
```
AppName		名称		注册方式	OnLine 		机器地址 		操作
golang-jobs	golang执行器	自动注册 		查看 ( 1 ）   
```
View->Registered Nodes
```
http://127.0.0.1:9999
```
### Add Tasks
Task management - > Add (note that the bean mode is used, and the name of jobhandler is the same as that of regtask)
```
1	测试panic	BEAN：task.panic	* 0 * * * ?	admin	STOP	
2	测试耗时任务	BEAN：task.test2	* * * * * ?	admin	STOP	
3	测试golang	BEAN：task.test		* * * * * ?	admin	STOP
```

