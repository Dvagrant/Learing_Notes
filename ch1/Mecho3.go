package main

import (
	"fmt"
	"os"
	"strings"
)

// TO USE GO TEST
// JOIN 方法： 将数组中的元素与指定字符串拼接，比传统的 str1 + str2 效率要高，因为无需创建销毁字符串对象
func main(){
	fmt.Println(strings.Join(os.Args[1:],"，"))
}