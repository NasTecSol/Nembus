package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"sync"
	"unsafe"
)

// Global database path and runtime state
var (
	dbPathMu sync.RWMutex
	globalDBPath string
)

//export InitNembusMobile
func InitNembusMobile(cDbPath *C.char) *C.char {
	dbPath := C.GoString(cDbPath)
	
	dbPathMu.Lock()
	globalDBPath = dbPath
	dbPathMu.Unlock()

	res := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"status":  "initialized",
			"db_path": dbPath,
		},
	}
	return toCString(res)
}

//export FetchCompleteTenantMasterData
func FetchCompleteTenantMasterData(cTenantId *C.char, cCloudUrl *C.char) *C.char {
	tenantID := C.GoString(cTenantId)
	cloudURL := C.GoString(cCloudUrl)

	dbPathMu.RLock()
	activeDB := globalDBPath
	dbPathMu.RUnlock()

	resp := SyncTenantBackupData(tenantID, cloudURL, activeDB)
	return toCString(resp)
}

//export CallHandler
func CallHandler(cReqJSON *C.char) *C.char {
	reqStr := C.GoString(cReqJSON)
	var req struct {
		Handler string          `json:"handler"`
		Action  string          `json:"action"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal([]byte(reqStr), &req); err != nil {
		return toCString(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON payload: " + err.Error(),
		})
	}

	res := DispatchHandler(req.Handler, req.Action, req.Payload)
	return toCString(res)
}

//export FreeCString
func FreeCString(ptr *C.char) {
	if ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func toCString(v interface{}) *C.char {
	b, err := json.Marshal(v)
	if err != nil {
		errMap := map[string]interface{}{"success": false, "error": err.Error()}
		b, _ = json.Marshal(errMap)
	}
	return C.CString(string(b))
}

func main() {}