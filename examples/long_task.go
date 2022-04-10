package examples

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/gohutool/boot4go-xxljob-executor"
	"strconv"
	"time"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : long_task.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 15:30
* 修改历史 : 1. [2022/4/10 15:30] 创建文件 by NST
*/

func LongTimeTask(cxt context.Context, param *RunReq) (string, error) {
	rtn := 0
	num := 0
	times := 6 * 2
	interval := 10

	m := make(map[string]string)

	if json.Unmarshal([]byte(param.ExecutorParams), m) == nil {
		if n, err := strconv.Atoi(m["num"]); err == nil {
			num = n
		}
		if n, err := strconv.Atoi(m["times"]); err == nil {
			times = n
		}
		if n, err := strconv.Atoi(m["interval"]); err == nil {
			interval = n
		}
	}

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

}
