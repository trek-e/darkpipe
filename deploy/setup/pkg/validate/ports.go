package validate

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

// ValidatePort checks if a TCP port is accessible
func ValidatePort(hostname string, port int, timeout time.Duration) error {
	address := net.JoinHostPort(hostname, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("port %d not accessible: %w", port, err)
	}
	defer conn.Close()

	return nil
}

// CheckLocalPorts checks which ports are already in use locally
func CheckLocalPorts(ports []int) ([]int, error) {
	inUse := []int{}

	for _, port := range ports {
		address := net.JoinHostPort("localhost", fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			// Port is in use
			conn.Close()
			inUse = append(inUse, port)
		}
	}

	return inUse, nil
}

// DetectRAM returns the total system memory in bytes
func DetectRAM() (uint64, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// On most systems, the total system memory is approximated
	// Note: runtime.MemStats doesn't provide total system RAM directly
	// This is a limitation - we'd need platform-specific code for accurate detection
	// For now, we return an approximation based on allocated memory
	// A better implementation would use syscall on Linux or similar platform-specific calls

	// For cross-platform compatibility, we'll return 0 and note in the error
	// that platform-specific detection is needed
	return 0, fmt.Errorf("platform-specific RAM detection not yet implemented - use manual check")
}
