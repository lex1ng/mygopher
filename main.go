package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Counter struct {
	addChan chan int
	getChan chan chan int
}

func NewCounter() *Counter {

	addChan := make(chan int)
	getChan := make(chan chan int)

	go func() {
		num := 0
		for {
			select {
			case val := <-addChan:
				num += val
			case ch := <-getChan:
				ch <- num
			}
		}
	}()

	return &Counter{
		addChan: addChan,
		getChan: getChan,
	}
}
func (c *Counter) Add(val int) {
	c.addChan <- val
}

func (c *Counter) Get() int {
	ch := make(chan int)
	c.getChan <- ch
	return <-ch
}

var client = &http.Client{
	Timeout: 5 * time.Second,
}

var (
	// 命令行参数
	input     = flag.String("input", "data.csv", "input file")
	workerNum = flag.Int("worker", 100, "worker num")
)

func init() {
	flag.Parse()
}

func check(url string) bool {
	if url == "" {
		return false
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func getFileWriter(filename string) (*csv.Writer, *os.File, error) {
	//defer close(channel)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("create file error:", err)
		return nil, file, err
	}
	writer := csv.NewWriter(file)

	heads := []string{"", "平台", "作者", "级别", "发布链接"}
	if err := writer.Write(heads); err != nil {
		fmt.Println("write to csv error:", err)
		return nil, file, err
	}
	return writer, file, nil
}

// 读取data csv文件

func main() {
	startTime := time.Now()
	defer func() {
		fmt.Println("used time:", time.Since(startTime))
	}()
	file, err := os.Open(*input)
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("read file error: %v", err)
		return
	}
	totalRecords := len(records) - 1

	fmt.Printf("Total records to process: %d\n", totalRecords)
	counter := NewCounter()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			current := counter.Get()
			ratio := float64(current) / float64(totalRecords)
			toPrint := strings.Repeat("*", int(100*ratio)) + strings.Repeat(" ", int(100*(1-ratio)))
			fmt.Printf("[%s] %d/%d\n", toPrint, current, 3555)
		}
	}()

	taskCh := make(chan []string)
	goodCh := make(chan []string)
	badCh := make(chan []string)
	// read data
	_, err = file.Seek(0, 0)
	if err != nil {
		fmt.Println("seek file error:", err)
		return
	}
	go func() {
		defer close(taskCh)
		fmt.Println("readData....")

		reader := csv.NewReader(file)

		for {
			record, err := reader.Read()
			if err != nil {
				if err == csv.ErrFieldCount || err == csv.ErrTrailingComma {
					continue // 忽略 CSV 解析错误
				}
				fmt.Printf("read data error: %v\n", err)
				break
			}
			if record[0] == "" {
				continue
			}
			taskCh <- record
		}
	}()

	// start
	var wg sync.WaitGroup
	for i := 0; i < *workerNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				if check(task[len(task)-1]) {
					goodCh <- task
				} else {
					badCh <- task
				}
				counter.Add(1)
			}
		}()
	}

	// save good result
	goodWriter, goodFile, gooderr := getFileWriter("good.csv")
	if gooderr != nil {
		fmt.Println("create good.csv error:", gooderr)
		return
	}
	defer func() {
		defer goodFile.Close()
		defer goodWriter.Flush()
	}()

	// save bad result
	badWriter, badFile, baderr := getFileWriter("bad.csv")
	if baderr != nil {
		fmt.Println("create bad.csv error", baderr)
	}
	defer func() {
		defer badFile.Close()
		defer badWriter.Flush()
	}()

	goodDone := make(chan struct{})
	badDone := make(chan struct{})

	go func() {
		for record := range goodCh {
			if err := goodWriter.Write(record); err != nil {
				fmt.Println("write to csv error:", err)
				return
			}
		}
		close(goodDone)
	}()

	go func() {
		// 如果在这里关闭ch，永远不回退出，因为ch没关闭，一直阻塞
		for record := range badCh {
			if err := badWriter.Write(record); err != nil {
				fmt.Println("write to csv error:", err)
				return
			}
		}
		close(badDone)
	}()

	// 所有task被处理完
	wg.Wait()
	// 关闭good和 bad，意思是不会再有新数据进去，两个携程读完之后，会自己退出
	close(goodCh)
	close(badCh)

	// 等待两个写携程处理完
	<-badDone
	<-goodDone
}
