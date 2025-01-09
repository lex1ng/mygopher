package checker

import (
	"main/pkg/result"
	"net/http"
	"sync"
	"time"
)

type Checker struct {
	sync.Mutex
	count   int
	total   int
	records [][]string
	good    *result.Result
	bad     *result.Result
}

func NewChecker(records [][]string) *Checker {
	return &Checker{
		count:   0,
		total:   len(records),
		records: records,
		good: &result.Result{
			Results: make([][]string, 0),
		},
		bad: &result.Result{
			Results: make([][]string, 0),
		},
	}
}
func (c *Checker) GetGoodResult() [][]string {
	return c.good.Results
}

func (c *Checker) GetBadResult() [][]string {
	return c.bad.Results
}
func (c *Checker) next() int {
	c.Lock()
	defer c.Unlock()
	c.count++
	return c.count
}

func check(url string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Head(url)
	if err != nil {
		// 链接不可达（网络问题、域名无法解析等）
		//fmt.Printf("url:%s, is unreachable %v \n ", url, err)
		return false
	}
	defer resp.Body.Close()

	// 判断状态码是否在 2xx 范围
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
func (c *Checker) StartWorker(idx int) {

	for c.count < c.total-1 {
		nextIdx := c.next()
		if nextIdx >= c.total {
			break
		}

		//fmt.Printf("worker %d is running with %d\n", idx, nextIdx)
		record := c.records[nextIdx]

		url := record[len(record)-1]
		if url == "" {
			c.bad.Append(record)
			continue
		}
		if check(url) {
			c.good.Append(record)
		} else {
			c.bad.Append(record)
		}

	}
}
