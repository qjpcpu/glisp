package repl

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func GetHistory() *History {
	h := &History{RWMutex: new(sync.RWMutex)}
	h.sync()
	return h
}

type History struct {
	*sync.RWMutex
	List []string
}

func (h *History) Get() []string {
	h.RLock()
	defer h.RUnlock()
	ret := make([]string, len(h.List))
	copy(ret, h.List)
	return ret
}

func (h *History) Append(v string) {
	h.Lock()
	defer h.Unlock()
	if v != `` {
		h.List = append(h.List, v)
	}
}

func (h *History) sync() {
	file := filepath.Join(os.TempDir(), `.glisp_history`)
	data, _ := ioutil.ReadFile(file)
	h.Lock()
	json.Unmarshal(data, &h.List)
	h.Unlock()
	go func() {
		timer := time.NewTicker(time.Second * 3)
		for range timer.C {
			h.Lock()
			if len(h.List) > 1000 {
				h.List = h.List[len(h.List)-1000:]
			}
			data, _ = json.Marshal(h.List)
			ioutil.WriteFile(file, data, 0755)
			h.Unlock()
		}
	}()
}
