package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadData(filename string, taskChan chan []string) {
	defer close(taskChan)
	fmt.Println("readData....")
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("open file error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err != nil {
			fmt.Printf("read data error: %v\n", err)
			break
		}
		if record[0] == "" {
			continue
		}
		taskChan <- record
	}

}

func ToCSV(channel chan []string, filename string, done chan struct{}) {
	defer close(done)
	//defer close(channel)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("create file error:", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	heads := []string{"", "平台", "作者", "级别", "发布链接"}
	if err := writer.Write(heads); err != nil {
		fmt.Println("write to csv error:", err)
		return
	}

	for record := range channel {
		if err := writer.Write(record); err != nil {
			fmt.Println("write to csv error:", err)
			return
		}
	}
}
