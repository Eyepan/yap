package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Eyepan/yap/src/utils"
)

func HandleList() {
	lockBin, err := utils.ReadLock()
	if err != nil {
		log.Fatalf("failed to read lockfile %v", err)
	}
	jsonData, err := json.MarshalIndent(lockBin, "", "\t")
	if err != nil {
		log.Fatalf("failed to parse lockfile into json %v", err)
	}
	fmt.Println(string(jsonData))
}
