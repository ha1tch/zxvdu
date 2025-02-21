package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// DrawCommand represents a drawing or control instruction.
type DrawCommand struct {
	Cmd    string   // Command name
	Params []int    // Numeric parameters
	Mode   string   // Mode flags ("S"/"F" for shapes, "flip"/"layer" for paint)
	Str    string   // String data (for tex commands)
	Conn   net.Conn // Connection for responses
}

// Command channel for passing commands from network to main loop
var commandChan = make(chan DrawCommand, 100)

// Event handling
var (
	eventConns   = make([]net.Conn, 0)
	eventConnsMu sync.Mutex
)

// startDrawingCommandServer listens on a TCP port for drawing commands.
func startDrawingCommandServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("ERROR 0001 : Error starting drawing command server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Drawing command server listening on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("ERROR 0001 : Error accepting drawing command connection:", err)
			continue
		}
		go handleDrawingCommandConn(conn)
	}
}

// startEventServer listens for event connections on a TCP port.
func startEventServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("ERROR 0001 : Error starting event server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Event server listening on", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("ERROR 0001 : Error accepting event connection:", err)
			continue
		}
		eventConnsMu.Lock()
		eventConns = append(eventConns, conn)
		eventConnsMu.Unlock()
		fmt.Println("New event client connected:", conn.RemoteAddr())
	}
}

// sendEvent broadcasts an event string to all connected event clients.
func sendEvent(event string) {
	eventConnsMu.Lock()
	defer eventConnsMu.Unlock()
	for i := 0; i < len(eventConns); i++ {
		_, err := fmt.Fprintln(eventConns[i], event)
		if err != nil {
			eventConns[i].Close()
			eventConns = append(eventConns[:i], eventConns[i+1:]...)
			i--
		}
	}
}

// handleDrawingCommandConn reads commands from a TCP connection.
func handleDrawingCommandConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		cmd, err := parseCommand(line)
		if err != nil {
			fmt.Fprintln(conn, err)
			continue
		}
		cmd.Conn = conn
		commandChan <- cmd
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("ERROR 0001 : Error reading from drawing command connection:", err)
	}
}

// processQuery handles query commands and sends a reply.
func processQuery(conn net.Conn, cmd string) {
	var response string
	switch cmd {
	case "colour":
		response = fmt.Sprintf("%d %d %d", defaultInk, defaultPaper, boolToInt(defaultBright))
	case "graphics":
		response = fmt.Sprintf("%d", graphicsMult)
	case "zoom":
		response = fmt.Sprintf("%d", zoomFactor)
	case "ink":
		response = fmt.Sprintf("%d", defaultInk)
	case "paper":
		response = fmt.Sprintf("%d", defaultPaper)
	case "bright":
		response = fmt.Sprintf("%d", boolToInt(defaultBright))
	case "host":
		response = "zxvdu v1.0"
	case "flip":
		response = fmt.Sprintf("%d", activeFlipBuffer)
	case "layer":
		response = fmt.Sprintf("%d", activeLayerBuffer)
	case "paint":
		response = currentDrawingMode
	case "paint_target":
		response = currentTarget
	case "eraser":
		response = fmt.Sprintf("%d", boolToInt(eraserMode))
	default:
		response = "unknown query"
	}
	fmt.Fprintln(conn, response)
}

// processCommands consumes commands from the command channel.
func processCommands() {
	for {
		select {
		case cmd := <-commandChan:
			if cmd.Mode == "query" {
				processQuery(cmd.Conn, cmd.Cmd)
				continue
			}

			switch cmd.Cmd {
			case "cls":
				handleCLS()
			case "flip":
				handleFlip(cmd)
			case "layer":
				handleLayer(cmd)
			case "paint":
				handlePaint(cmd)
			case "paint_target":
				handlePaintTarget(cmd)
			case "paint_copy":
				handlePaintCopy(cmd)
			case "ink":
				if len(cmd.Params) == 1 {
					defaultInk = cmd.Params[0]
					if currentDrawingMode == "layer" {
						eraserMode = false
					}
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
					if currentDrawingMode == "layer" {
						eraserMode = false
					}
				}
			case "graphics":
				handleGraphics(cmd)
			case "zoom":
				handleZoom(cmd)
			case "eraser":
				if currentDrawingMode == "layer" {
					eraserMode = true
				}
			case "tex":
				handleTexCommand(cmd)
			default:
				updateActiveBuffer(cmd)
			}
		default:
			return
		}
	}
}

func handlePaintTarget(cmd DrawCommand) {
	if cmd.Mode == "onscreen" || cmd.Mode == "offscreen" {
		currentTarget = cmd.Mode
	}
}

func handlePaintCopy(cmd DrawCommand) {
	if err := copyBufferFromOffscreen(cmd.Mode, cmd.Params[0], cmd.Params[1]); err != nil {
		if cmd.Conn != nil {
			fmt.Fprintln(cmd.Conn, "ERROR 0030 :", err)
		}
	}
}

func handleFlip(cmd DrawCommand) {
	if len(cmd.Params) == 0 {
		if activeFlipBuffer == 0 {
			activeFlipBuffer = 1
		} else {
			activeFlipBuffer = 0
		}
	} else if len(cmd.Params) == 1 {
		if cmd.Params[0] >= 0 && cmd.Params[0] < numFlipBuffers {
			activeFlipBuffer = cmd.Params[0]
		}
	}
}

func handleLayer(cmd DrawCommand) {
	if len(cmd.Params) == 0 {
		if activeLayerBuffer == 0 {
			activeLayerBuffer = 1
		} else {
			activeLayerBuffer = 0
		}
	} else if len(cmd.Params) == 1 {
		if cmd.Params[0] >= 0 && cmd.Params[0] < numLayerBuffers {
			activeLayerBuffer = cmd.Params[0]
		}
	}
}

func handlePaint(cmd DrawCommand) {
	if strings.ToLower(cmd.Mode) == "flip" {
		currentDrawingMode = "flip"
		eraserMode = false
	} else if strings.ToLower(cmd.Mode) == "layer" {
		currentDrawingMode = "layer"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}