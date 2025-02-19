package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metrics"
)

var monitorFile *os.File
var globalParam *logKV
var blockchainObj *blockchain.BlockChain

func init() {
	uid := uuid.New()
	globalParam = &logKV{param: make(map[string]interface{})}
	SetGlobalParam("UID", uid.String())
	commitID := os.Getenv("commit")
	if commitID == "" {
		commitID = "NA"
	}
	SetGlobalParam("CommitID", commitID)

	go func() {
		ticker := time.NewTicker(40 * time.Second)
		idle0, total0 := common.GetCPUSample()
		var m runtime.MemStats
		for range ticker.C {
			if blockchainObj == nil {
				time.Sleep(time.Second)
				continue
			}
			l := NewLog()
			idle1, total1 := common.GetCPUSample()
			idleTicks := float64(idle1 - idle0)
			totalTicks := float64(total1 - total0)
			cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
			runtime.ReadMemStats(&m)

			// Disk usage (cross-platform)
			diskUsage := getDiskUsage("/data") // Use appropriate path for Windows or Linux
			l.Add("DISK_USAGE", diskUsage)
			l.Add("CPU_USAGE", fmt.Sprintf("%.2f", cpuUsage), "MEM_USAGE", m.Sys>>20)
			idle0, total0 = common.GetCPUSample()
			l.Write()
		}
	}()
}

type logKV struct {
	param map[string]interface{}
	sync.RWMutex
}

func SetGlobalParam(p ...interface{}) {
	globalParam.Add(p...)
}

func SetBlockChainObj(obj *blockchain.BlockChain) {
	blockchainObj = obj
}

func NewLog(p ...interface{}) *logKV {
	nl := (&logKV{param: make(map[string]interface{})}).Add(p...)
	globalParam.RLock()
	for k, v := range globalParam.param {
		nl.param[k] = v
	}
	globalParam.RUnlock()
	return nl
}

func (s *logKV) Add(p ...interface{}) *logKV {
	if len(p) == 0 || len(p)%2 != 0 {
		return s
	}
	s.Lock()
	defer s.Unlock()
	for i, v := range p {
		if i%2 == 0 {
			s.param[v.(string)] = p[i+1]
		}
	}
	return s
}

func (s *logKV) Write() {
	s.RLock()
	defer s.RUnlock()
	s.param["Time"] = fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
	b, _ := json.Marshal(s.param)
	if v, ok := s.param["MINING_PUBKEY"]; !ok || v == "" {
		return
	}

	go func() {
		monitorEP := os.Getenv("MONITOR")
		if monitorEP != "" {
			req, err := http.NewRequest(http.MethodPost, monitorEP, bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				metrics.IncLogger.Log.Debug("Create Request failed with err: ", err)
				return
			}
			ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
			defer cancel()
			req = req.WithContext(ctx)
			client := &http.Client{}
			client.Do(req)
		}
	}()
}

func getMethodName(depthList ...int) (string, string, string) {
	var depth int
	if depthList == nil {
		depth = 1
	} else {
		depth = depthList[0]
	}
	function, file, line, _ := runtime.Caller(depth)
	r, _ := regexp.Compile("([^/]*$)")
	r1, _ := regexp.Compile("/([^/]*$)")
	return r.FindStringSubmatch(runtime.FuncForPC(function).Name())[1], r1.FindStringSubmatch(file)[1], strconv.Itoa(line)
}
