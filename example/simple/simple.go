package main

import (
	"flag"
	"fmt"
	"live/pkg"
	"net"
	"os"
	"os/signal"
	"syscall"
	_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据
	_ "net/http/pprof"                // 必须，引入 pprof 模块
)

var (
	mqPort = flag.Int("mq_port", 1883, "mqtt port")
)

func main() {

	go func() {
		// terminal: $ go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
		// web:
		// 1、http://localhost:8081/ui
		// 2、http://localhost:6060/debug/charts
		// 3、http://localhost:6060/debug/pprof
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	address := fmt.Sprintf("0.0.0.0:%d", *mqPort)

	netAddr2, _ := net.ResolveTCPAddr("tcp", address)
	pkg.NewServer(netAddr2)

	c := make(chan os.Signal)
	//监听指定信号
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)

	s := <-c
	//收到信号后的处理，这里只是输出信号内容，可以做一些更有意思的事
	fmt.Println("get signal:", s)

}
