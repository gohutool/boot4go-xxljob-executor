package examples

import (
	"context"
	"errors"
	"fmt"
	. "github.com/gohutool/boot4go-xxljob-executor"
	. "github.com/gohutool/boot4go-xxljob-executor/utils"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : task.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 15:29
* 修改历史 : 1. [2022/4/10 15:29] 创建文件 by NST
*/

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
