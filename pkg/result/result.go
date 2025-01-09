package result

import "sync"

type Result struct {
	sync.Mutex
	Results [][]string
}

func (r *Result) Append(record []string) {
	r.Lock()
	defer r.Unlock()
	r.Results = append(r.Results, record)
}
