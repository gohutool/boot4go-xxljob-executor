# boot4go-xxljob-executor
golang for xxl-job executor support

很多公司java与go开发共存，java中有xxl-job做为任务调度引擎，为此也出现了go执行器(客户端)，使用起来比较简单：
# 支持
```	
1.执行器注册
2.耗时任务取消
3.任务注册，像写http.Handler一样方便
4.任务panic处理
5.阻塞策略处理
6.任务完成支持返回执行备注
7.任务超时取消 (单位：秒，0为不限制)
8.失败重试次数(在参数param中，目前由任务自行处理)
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
	log4go.LoggerManager.InitWithDefaultConfig()
	logger := log4go.LoggerManager.GetLogger("gohutool.executor4go.examples")

	exec := NewExecutor(
		ServerAddr("http://192.106.56.101/xxl-job-admin"),
		AccessToken(""),            //请求令牌(默认为空)
		ExecutorIp("192.106.56.1"), //可自动获取
		ExecutorPort("9999"),       //默认9999（非必填）
		RegistryKey("golang-jobs"), //执行器名称
		SetLogger(logger),          //自定义日志
	)
	
		exec.Init()
	//设置日志查看handler
	exec.SetLogHandler(func(req *LogReq) *LogRes {
		return &LogRes{Code: 200, Msg: "", Content: LogResContent{
			FromLineNum: req.FromLineNum,
			ToLineNum:   2,
			LogContent:  "这个是自定义日志handler",
			IsEnd:       true,
		}}
	})
	//注册任务handler
	exec.RegTask("task.simple", SimpleTask)
	exec.RegTask("task.longTime", LongTimeTask)
	exec.RegTask("task.panic", PanicTask)
	exec.RegTask("task.exception", ExceptionTask)

	logger.Info(exec.Run())
}

```

启动日志

```
[18:40:58 CST 2022/04/10 583] [INFO][gohutool.executor4go.examples] (github.com/gohutool/boot4go-xxljob-executor.(*Executor).Init:112) executor4go-v1.0.0 192.106.56.1:9999
[18:40:58 CST 2022/04/10 583] [INFO][gohutool.executor4go.examples] (github.com/gohutool/boot4go-xxljob-executor.(*RestFulExecutor).Run:49) Starting server at 192.106.56.1:9999
[18:41:33 CST 2022/04/10 252] [EROR][gohutool.executor4go.examples] (github.com/gohutool/boot4go-xxljob-executor.(*RestFulExecutor).Registry.func1:84) 执行器注册失败1:Post "http
://192.106.56.101/xxl-job-admin/api/registry": dial tcp 192.106.56.101:80: connectex: A connection attempt failed because the connected party did not properly respond after a pe
riod of time, or established connection failed because connected host has failed to respond.

```

```
    func SimpleTask(cxt context.Context, param *RunReq) (string, error) {
        fmt.Println("test one task" + param.ExecutorHandler + " param：" +
        param.ExecutorParams + " log_id:" + Int64ToStr(param.LogID))
        return "test done", nil
    }
    
    func PanicTask(cxt context.Context, param *RunReq) (string, error) {
        panic("test one task" + param.ExecutorHandler + " param：" +
        param.ExecutorParams + " log_id:" + Int64ToStr(param.LogID))
    }
    
    func ExceptionTask(cxt context.Context, param *RunReq) (string, error) {
        return "", errors.New("test one task" + param.ExecutorHandler +
        " param：" + param.ExecutorParams + " log_id:" + Int64ToStr(param.LogID))
    }
```

```
func LongTimeTask(cxt context.Context, param *RunReq) (string, error) {
    
    ....

	for {
		select {
		case <-cxt.Done():
			fmt.Println("task" + param.ExecutorHandler + "被手动终止")
			return fmt.Sprintf("task"+param.ExecutorHandler+"被手动终止, 执行次数%d 返回值为%d", num, rtn), nil
		default:
			num++

			rtn++

			time.Sleep(time.Duration(interval) * time.Second)

			fmt.Printf("test one task"+param.ExecutorHandler+" param："+param.ExecutorParams+" 执行次数%d 值为%d\n",
				num, rtn)

			if num > times {
				fmt.Printf("test one task"+param.ExecutorHandler+
					" param："+param.ExecutorParams+"执行完毕, 执行次数%d 值为%d！\n", num, rtn)
				return fmt.Sprintf("test one task"+param.ExecutorHandler+
					" param："+param.ExecutorParams+"执行完毕, 执行次数%d 值为%d！\n", num, rtn), nil
			}
		}
	}
	
	....

}
```

# 示例项目
github.com/gohutool/boot4go-xxljob-executor/examples/

# xxl-job-admin配置
### 添加执行器
执行器管理->新增执行器,执行器列表如下：
```
AppName		名称		注册方式	OnLine 		机器地址 		操作
golang-jobs	golang执行器	自动注册 		查看 ( 1 ）   
```
查看->注册节点
```
http://192.168.56.1:9999
```
### 添加任务
任务管理->新增(注意，使用BEAN模式，JobHandler与RegTask名称一致)
```
1	测试简单任务	BEAN：task.simple	    * 0 * * * ?	admin	STOP	
2	测试耗时任务	BEAN：task.longTime	    * * * * * ?	admin	STOP	
3	测试异常任务	BEAN：task.panic		* * * * * ?	admin	STOP
4	测试错误任务	BEAN：task.exception	* * * * * ?	admin	STOP
```

