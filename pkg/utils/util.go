package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

func ReadData(filename string) ([][]string, error) {
	fmt.Println("readData....")

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	return records[1:], nil

}

func ToCSV(results [][]string, filename string) {
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

	for _, record := range results {
		if err := writer.Write(record); err != nil {
			fmt.Println("write to csv error:", err)
			return
		}
	}
}
