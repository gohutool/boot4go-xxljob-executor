package utils

import "strconv"

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : util.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 12:34
* 修改历史 : 1. [2022/4/10 12:34] 创建文件 by NST
*/

// Int64ToStr int64 to str
func Int64ToStr(i int64) string {
	return strconv.FormatInt(i, 10)
}
