# result
go小白， 读取文件和检查链接可用 借助了AI。

程序耗时 10.5s，主要因为开了100个协程。

good和bad 有出入，主要因为检查是否可用的方法有问题

main 可执行文件是在MAC环境下编译的可执行文件
# start 
```shell
go build
# worker 是go协助数量
./main --worker 100 
```