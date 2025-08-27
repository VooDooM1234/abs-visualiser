package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func Encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// STOP STEALIN MY PORTS
func KillPort(port string) error {
	psCommand := fmt.Sprintf(`$connections = Get-NetTCPConnection -LocalPort %s -ErrorAction SilentlyContinue;
		foreach ($c in $connections) {
			Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
		}`, port)

	cmd := exec.Command("powershell", "-Command", psCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// generic check for mservice response of {"status": "success/fail"}
func CheckFailureResponse(body []byte, logger *log.Logger) error {
	var genericResp map[string]interface{}
	if err := json.Unmarshal(body, &genericResp); err != nil {
		logger.Printf("Failed to parse JSON: %v", err)
		return fmt.Errorf("invalid response format")
	}

	if status, ok := genericResp["status"]; ok && status == "failed" {
		message := "Unknown error"
		if msg, exists := genericResp["message"]; exists {
			message = fmt.Sprintf("%v", msg)
		}
		logger.Printf("Backend returned failure: %s", message)
		return fmt.Errorf(message)
	}

	return nil
}
