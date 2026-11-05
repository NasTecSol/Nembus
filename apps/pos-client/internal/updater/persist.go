package updater

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type launchArgsFile struct {
	Args []string `json:"args"`
}

// PersistLaunchArgs records CLI args (e.g. prod/stg) so an in-place update restart
// preserves the same runtime profile.
func PersistLaunchArgs(args []string) {
	if len(args) == 0 {
		return
	}
	dir, err := nembusDir()
	if err != nil {
		log.Printf("[UPDATER] Could not persist launch args: %v", err)
		return
	}
	path := filepath.Join(dir, "launch_args.json")
	payload, err := json.Marshal(launchArgsFile{Args: args})
	if err != nil {
		return
	}
	if err := os.WriteFile(path, payload, 0644); err != nil {
		log.Printf("[UPDATER] Could not write launch args: %v", err)
	}
}

// LoadLaunchArgs returns persisted CLI args for the next process start.
func LoadLaunchArgs() []string {
	dir, err := nembusDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(dir, "launch_args.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var file launchArgsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil
	}
	return file.Args
}

func nembusDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".nembus")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
