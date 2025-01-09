package main

import (
	"flag"
	"fmt"
	"main/pkg/checker"
	"main/pkg/utils"
	"sync"
	"time"
)

var (
	// 命令行参数
	input     = flag.String("input", "data.csv", "input file")
	workerNum = flag.Int("worker", 100, "worker num")
)

func init() {
	flag.Parse()
}

// 读取data csv文件

func main() {
	startTime := time.Now()

	// read data
	records, err := utils.ReadData(*input)
	if err != nil {
		fmt.Println("readData err:", err)
		return
	}

	// init checker
	check := checker.NewChecker(records)

	// start
	var wg sync.WaitGroup
	for i := 0; i < *workerNum; i++ {
		wg.Add(1)
		tmp := i
		go func() {
			check.StartWorker(tmp)
			defer wg.Done()
		}()
	}
	wg.Wait()

	// save result
	utils.ToCSV(check.GetGoodResult(), "good.csv")
	utils.ToCSV(check.GetBadResult(), "bad.csv")

	usedTime := time.Since(startTime)
	fmt.Println("used time:", usedTime)
}
