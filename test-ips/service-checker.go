package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// ServiceResult stores the result of a service availability check
type ServiceResult struct {
	Address  string
	Available bool
	Error    string
	Duration time.Duration
}

func main() {
	// Define command line flag
	servicesFlag := flag.String("services", "", "Comma-separated list of IP:port services to check (e.g., 192.168.1.1:80,10.0.0.1:443)")
	timeoutFlag := flag.Int("timeout", 5, "Connection timeout in seconds")
	
	// Parse command line arguments
	flag.Parse()
	
	// Validate services input
	if *servicesFlag == "" {
		fmt.Println("Error: No services specified")
		fmt.Println("Usage: service-checker -services=IP1:port1,IP2:port2,...")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	// Parse services list
	serviceAddresses := strings.Split(*servicesFlag, ",")
	
	// Set timeout duration
	timeout := time.Duration(*timeoutFlag) * time.Second
	
	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Create a channel to collect results
	results := make(chan ServiceResult, len(serviceAddresses))
	
	// Check each service in parallel
	for _, address := range serviceAddresses {
		address = strings.TrimSpace(address)
		if address == "" {
			continue
		}
		
		wg.Add(1)
		go checkService(address, timeout, results, &wg)
	}
	
	// Wait for all checks to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Process and print results
	unavailableCount := 0
	
	fmt.Println("Service Availability Check Results:")
	fmt.Println("==================================")
	
	for result := range results {
		if result.Available {
			fmt.Printf("✅ %s - Available (%.2f seconds)\n", result.Address, result.Duration.Seconds())
		} else {
			unavailableCount++
			fmt.Printf("❌ %s - Unavailable: %s (%.2f seconds)\n", result.Address, result.Error, result.Duration.Seconds())
		}
	}
	
	fmt.Println("==================================")
	fmt.Printf("Summary: %d services checked, %d unavailable\n", len(serviceAddresses), unavailableCount)
	
	// Return exit code 1 if any services are unavailable
	if unavailableCount > 0 {
		os.Exit(1)
	}
}

// checkService attempts to connect to a service and reports the result
func checkService(address string, timeout time.Duration, results chan<- ServiceResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	startTime := time.Now()
	
	// Validate address format
	if !strings.Contains(address, ":") {
		results <- ServiceResult{
			Address:   address,
			Available: false,
			Error:     "invalid address format (should be IP:port)",
			Duration:  time.Since(startTime),
		}
		return
	}
	
	// Attempt to establish TCP connection
	conn, err := net.DialTimeout("tcp", address, timeout)
	
	// Create result object
	result := ServiceResult{
		Address:   address,
		Duration:  time.Since(startTime),
	}
	
	// Check connection result
	if err != nil {
		result.Available = false
		result.Error = err.Error()
	} else {
		result.Available = true
		conn.Close()
	}
	
	// Send result to channel
	results <- result
}