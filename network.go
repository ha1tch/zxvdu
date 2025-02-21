package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Command channel for passing commands from network to main loop
var commandChan = make(chan DrawCommand, 100)

// Event handling
var (
	eventConns   = make([]net.Conn, 0)
	eventConnsMu sync.Mutex
)

// startDrawingCommandServer listens on a TCP port for drawing commands
func startDrawingCommandServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting drawing command server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Drawing command server listening on", addr)
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting drawing command connection:", err)
			continue
		}
		go handleDrawingCommandConn(conn)
	}
}

// startEventServer listens for event connections on a TCP port
func startEventServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting event server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Event server listening on", addr)
	
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting event connection:", err)
			continue
		}
		eventConnsMu.Lock()
		eventConns = append(eventConns, conn)
		eventConnsMu.Unlock()
		fmt.Println("New event client connected:", conn.RemoteAddr())
	}
}

// sendEvent broadcasts an event string to all connected event clients
func sendEvent(event string) {
	eventConnsMu.Lock()
	defer eventConnsMu.Unlock()
	
	// Create new slice for active connections
	activeConns := make([]net.Conn, 0, len(eventConns))
	
	// Send to all connections, collecting active ones
	for _, conn := range eventConns {
		_, err := fmt.Fprintln(conn, event)
		if err == nil {
			activeConns = append(activeConns, conn)
		} else {
			conn.Close()
		}
	}
	
	// Replace connection list with active ones
	eventConns = activeConns
}

// handleDrawingCommandConn reads commands from a TCP connection
func handleDrawingCommandConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	
	for scanner.Scan() {
		line := scanner.Text()
		cmd, err := parseCommand(line)
		if err != nil {
			fmt.Fprintln(conn, "ERROR 0020 :", err)
			continue
		}
		
		// Handle queries directly
		if cmd.Mode == "query" {
			response := processQuery(cmd.Cmd)
			fmt.Fprintln(conn, response)
			continue
		}

		// Handle texture operations that need immediate response
		if isTextureOperation(cmd) {
			handleTextureOperation(cmd, conn)
			continue
		}

		// Send other commands to main loop
		select {
		case commandChan <- cmd:
			// Command sent successfully
		default:
			fmt.Fprintln(conn, "ERROR 0033 : server busy, try again later")
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from drawing command connection:", err)
	}
}

// isTextureOperation checks if a command needs immediate texture handling
func isTextureOperation(cmd DrawCommand) bool {
	return cmd.Cmd == "tex" || (cmd.Cmd == "rect" && strings.EqualFold(cmd.Mode, "T"))
}

// handleTextureOperation processes texture-related commands and sends responses
func handleTextureOperation(cmd DrawCommand, conn net.Conn) {
	var slot int
	var err error

	if cmd.Cmd == "tex" {
		slot, err = handleTexCommand(cmd)
	} else if cmd.Cmd == "rect" && strings.EqualFold(cmd.Mode, "T") {
		slot, err = handleTextureCapture(cmd)
	}

	if err != nil {
		switch err.Error() {
		case "no free texture slots":
			fmt.Fprintln(conn, "ERROR 0031 : no free texture slots available")
		case "invalid region bounds":
			fmt.Fprintln(conn, "ERROR 0030 : invalid capture region")
		case "no pixel data provided":
			fmt.Fprintln(conn, "ERROR 0021 : no pixel data provided")
		case "invalid texture number":
			fmt.Fprintln(conn, "ERROR 0022 : invalid texture number")
		case "invalid texture parameters":
			fmt.Fprintln(conn, "ERROR 0023 : invalid texture parameters")
		default:
			fmt.Fprintln(conn, "ERROR 0029 : texture operation failed:", err)
		}
		return
	}

	// Send successful texture slot number
	fmt.Fprintln(conn, slot)
}

// processCommands consumes commands from the command channel
func processCommands() {
	for {
		select {
		case cmd := <-commandChan:
			slot, err := executeCommand(cmd)
			if err != nil {
				fmt.Printf("command error: %v\n", err)
			}
			// Handle successful texture operations
			if slot >= 0 {
				fmt.Printf("texture operation successful: slot %d\n", slot)
			}
		default:
			return
		}
	}
}

// executeCommand processes a single drawing command
func executeCommand(cmd DrawCommand) (int, error) {
	switch cmd.Cmd {
	case "cls":
		if currentDrawingMode == "layer" {
			buffers.ClearLayer()
		} else {
			buffers.ClearFlip()
		}
		
	case "flip":
		n := 1 // default
		if len(cmd.Params) > 0 {
			n = cmd.Params[0]
		}
		if err := buffers.SwapFlip(n); err != nil {
			return -1, fmt.Errorf("flip error: %v", err)
		}
		
	case "layer":
		n := 1 // default
		if len(cmd.Params) > 0 {
			n = cmd.Params[0]
		}
		if err := buffers.SwapLayer(n); err != nil {
			return -1, fmt.Errorf("layer error: %v", err)
		}
		
	case "paint":
		if cmd.Mode == "flip" || cmd.Mode == "layer" {
			// Mode selection
			currentDrawingMode = cmd.Mode
		} else {
			// Buffer selection
			n := 0 // default to buffer 0
			if len(cmd.Params) > 0 {
				n = cmd.Params[0]
			}
			if err := buffers.SetActiveTarget(n); err != nil {
				return -1, fmt.Errorf("paint error: %v", err)
			}
		}
		
	case "ink":
		if len(cmd.Params) == 1 {
			defaultInk = cmd.Params[0]
		}
		
	case "paper":
		if len(cmd.Params) == 1 {
			defaultPaper = cmd.Params[0]
		}
		
	case "bright":
		if len(cmd.Params) == 1 {
			defaultBright = (cmd.Params[0] == 1)
		}
		
	case "colour":
		if len(cmd.Params) == 3 {
			defaultInk = cmd.Params[0]
			defaultPaper = cmd.Params[1]
			defaultBright = (cmd.Params[2] == 1)
		}
		
	default:
		// Drawing commands
		return updateActiveBuffer(buffers, cmd, currentDrawingMode == "layer")
	}

	return -1, nil
}