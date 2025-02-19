package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Global default drawing state.
var (
	// Default ink (foreground) and paper (background) are ZX Spectrum values (0–7).
	// Note: 0 is black, so the bright flag does not affect black.
	defaultInk   int  = 7 // default to white
	defaultPaper int  = 0 // default to black
	defaultBright bool = false
)

// DrawCommand represents a drawing or control instruction.
type DrawCommand struct {
	Cmd    string // "plot", "circle", "line", "lineto", "rect", "cls", "flip", "ink", "paper", "bright", "colour"
	Params []int  // parameters for drawing/control commands
}

// commandChan is used to send commands from network handlers to the main loop.
var commandChan = make(chan DrawCommand, 100)

// Global slice of active GUI event connections.
var (
	guiEventConns   = make([]net.Conn, 0)
	guiEventConnsMu sync.Mutex
)

func main() {
	// Start TCP servers for drawing commands and GUI events.
	go startDrawingCommandServer(":55555")
	go startGUIEventsServer(":55556")

	// Display dimensions.
	screenWidth := 1024
	screenHeight := 768

	// Initialize the window.
	rl.InitWindow(int32(screenWidth), int32(screenHeight), "VDU/Display Server")
	rl.SetTargetFPS(60)

	// Define the ZX Spectrum palette of 15 colours.
	// Non-bright colours: indices 0–7; bright variants: indices 8–14.
	palette := []rl.Color{
		rl.Black,                        // 0: Black (no bright variant)
		rl.NewColor(0, 0, 205, 255),       // 1: Blue
		rl.NewColor(205, 0, 0, 255),       // 2: Red
		rl.NewColor(205, 0, 205, 255),     // 3: Magenta
		rl.NewColor(0, 205, 0, 255),       // 4: Green
		rl.NewColor(0, 205, 205, 255),     // 5: Cyan
		rl.NewColor(205, 205, 0, 255),     // 6: Yellow
		rl.NewColor(205, 205, 205, 255),   // 7: White (normal)
		rl.NewColor(0, 0, 255, 255),       // 8: Bright Blue
		rl.NewColor(255, 0, 0, 255),       // 9: Bright Red
		rl.NewColor(255, 0, 255, 255),     // 10: Bright Magenta
		rl.NewColor(0, 255, 0, 255),       // 11: Bright Green
		rl.NewColor(0, 255, 255, 255),     // 12: Bright Cyan
		rl.NewColor(255, 255, 0, 255),     // 13: Bright Yellow
		rl.NewColor(255, 255, 255, 255),   // 14: Bright White
	}

	// Create up to 8 buffers for drawing.
	const numBuffers = 8
	buffers := make([][]DrawCommand, numBuffers)
	for i := 0; i < numBuffers; i++ {
		buffers[i] = []DrawCommand{}
	}
	activeBuffer := 0

	// Global current position for "lineto" commands.
	currX, currY := 0, 0

	// Main rendering loop.
	for !rl.WindowShouldClose() {
		// Process any new commands from the network.
		processCommands(&buffers, &activeBuffer, &currX, &currY)

		rl.BeginDrawing()
		// Clear the screen with Spectrum black.
		rl.ClearBackground(palette[0])

		// Replay drawing commands from the active buffer.
		for _, cmd := range buffers[activeBuffer] {
			switch cmd.Cmd {
			case "plot":
				// plot x y colorIndex (if -1, use effective ink)
				if len(cmd.Params) >= 3 {
					colorIndex := cmd.Params[2]
					if colorIndex == -1 {
						colorIndex = effectiveInkColor()
					}
					if colorIndex < len(palette) {
						rl.DrawPixel(int32(cmd.Params[0]), int32(cmd.Params[1]), palette[colorIndex])
					}
				}
			case "circle":
				// circle x y radius colorIndex (if -1, use effective ink)
				if len(cmd.Params) >= 4 {
					colorIndex := cmd.Params[3]
					if colorIndex == -1 {
						colorIndex = effectiveInkColor()
					}
					if colorIndex < len(palette) {
						rl.DrawCircle(int32(cmd.Params[0]), int32(cmd.Params[1]), float32(cmd.Params[2]), palette[colorIndex])
					}
				}
			case "line":
				// line x1 y1 x2 y2 colorIndex (if -1, use effective ink)
				if len(cmd.Params) >= 5 {
					colorIndex := cmd.Params[4]
					if colorIndex == -1 {
						colorIndex = effectiveInkColor()
					}
					if colorIndex < len(palette) {
						rl.DrawLine(int32(cmd.Params[0]), int32(cmd.Params[1]),
							int32(cmd.Params[2]), int32(cmd.Params[3]), palette[colorIndex])
					}
				}
			case "lineto":
				// lineto x y colorIndex (if -1, use effective ink); uses the global current position.
				if len(cmd.Params) >= 3 {
					colorIndex := cmd.Params[2]
					if colorIndex == -1 {
						colorIndex = effectiveInkColor()
					}
					if colorIndex < len(palette) {
						rl.DrawLine(int32(currX), int32(currY),
							int32(cmd.Params[0]), int32(cmd.Params[1]), palette[colorIndex])
						currX, currY = cmd.Params[0], cmd.Params[1]
					}
				}
			case "rect":
				// rect x y width height colorIndex (if -1, use effective paper)
				if len(cmd.Params) >= 5 {
					colorIndex := cmd.Params[4]
					if colorIndex == -1 {
						colorIndex = effectivePaperColor()
					}
					if colorIndex < len(palette) {
						rl.DrawRectangle(int32(cmd.Params[0]), int32(cmd.Params[1]),
							int32(cmd.Params[2]), int32(cmd.Params[3]), palette[colorIndex])
					}
				}
			}
		}

		// Example: broadcast a GUI event on left mouse click.
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()
			eventStr := fmt.Sprintf("mouse: %d,%d", int(mousePos.X), int(mousePos.Y))
			sendGUIEvent(eventStr)
		}

		rl.EndDrawing()
	}

	rl.CloseWindow()
}

// effectiveInkColor returns the computed ink (foreground) color index
// based on the defaultInk and defaultBright flag.
func effectiveInkColor() int {
	if defaultInk == 0 {
		return 0 // black is always black
	}
	if defaultBright {
		return defaultInk + 7
	}
	return defaultInk
}

// effectivePaperColor returns the computed paper (background) color index
// based on the defaultPaper and defaultBright flag.
func effectivePaperColor() int {
	if defaultPaper == 0 {
		return 0
	}
	if defaultBright {
		return defaultPaper + 7
	}
	return defaultPaper
}

// processCommands reads pending commands from commandChan and updates the buffers
// or global drawing state as needed.
func processCommands(buffers *[][]DrawCommand, activeBuffer *int, currX, currY *int) {
	for {
		select {
		case cmd := <-commandChan:
			switch cmd.Cmd {
			case "cls":
				// Clear the active buffer.
				(*buffers)[*activeBuffer] = []DrawCommand{}
			case "flip":
				if len(cmd.Params) == 0 {
					// Toggle between buffers 0 and 1.
					if *activeBuffer == 0 {
						*activeBuffer = 1
					} else {
						*activeBuffer = 0
					}
				} else if len(cmd.Params) == 1 {
					// Select the specified buffer if valid (0–7).
					if cmd.Params[0] >= 0 && cmd.Params[0] < len(*buffers) {
						*activeBuffer = cmd.Params[0]
					}
				}
			case "ink":
				// Set default ink colour.
				if len(cmd.Params) == 1 {
					defaultInk = cmd.Params[0]
				}
			case "paper":
				// Set default paper colour.
				if len(cmd.Params) == 1 {
					defaultPaper = cmd.Params[0]
				}
			case "bright":
				// Set brightness flag.
				if len(cmd.Params) == 1 {
					defaultBright = (cmd.Params[0] == 1)
				}
			case "colour":
				// Set ink, paper, and bright all at once.
				if len(cmd.Params) == 3 {
					defaultInk = cmd.Params[0]
					defaultPaper = cmd.Params[1]
					defaultBright = (cmd.Params[2] == 1)
				}
			default:
				// For drawing commands, append to the current active buffer.
				(*buffers)[*activeBuffer] = append((*buffers)[*activeBuffer], cmd)
			}
		default:
			return
		}
	}
}

// startDrawingCommandServer listens on the given address for drawing commands.
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

// handleDrawingCommandConn reads commands from a drawing command connection.
func handleDrawingCommandConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		cmd, err := parseCommand(line)
		if err != nil {
			fmt.Fprintf(conn, "Error parsing command: %v\n", err)
			continue
		}
		commandChan <- cmd
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from drawing command connection:", err)
	}
}

// parseCommand converts a text line into a DrawCommand.
// Supported commands:
//
//   Drawing commands:
//     - plot x y [colorIndex]        (if colorIndex is omitted or -1, uses default ink)
//     - circle x y radius [colorIndex] (if colorIndex is omitted or -1, uses default ink)
//     - line x1 y1 x2 y2 [colorIndex]  (if colorIndex is omitted or -1, uses default ink)
//     - lineto x y [colorIndex]        (if colorIndex is omitted or -1, uses default ink)
//     - rect x y width height [colorIndex] (if colorIndex is omitted or -1, uses default paper)
//
//   Control commands:
//     - cls                        (clears current active buffer)
//     - flip [bufferIndex]         (toggles or selects a buffer, 0–7)
//     - ink color                  (set default ink 0–7)
//     - paper color                (set default paper 0–7)
//     - bright 0|1                 (set brightness flag)
//     - colour ink paper bright    (set all three in one command)
func parseCommand(line string) (DrawCommand, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("empty command")
	}
	cmd := strings.ToLower(fields[0])
	var params []int
	for _, field := range fields[1:] {
		val, err := strconv.Atoi(field)
		if err != nil {
			return DrawCommand{}, fmt.Errorf("invalid parameter %q", field)
		}
		params = append(params, val)
	}

	switch cmd {
	case "plot":
		// Accepts 2 or 3 parameters; if 2 provided, use -1 for color.
		if len(params) == 2 {
			params = append(params, -1)
		}
		if len(params) != 3 {
			return DrawCommand{}, fmt.Errorf("plot requires 2 or 3 parameters: x y [colorIndex]")
		}
	case "circle":
		// Accepts 3 or 4 parameters.
		if len(params) == 3 {
			params = append(params, -1)
		}
		if len(params) != 4 {
			return DrawCommand{}, fmt.Errorf("circle requires 3 or 4 parameters: x y radius [colorIndex]")
		}
	case "line":
		// Accepts 4 or 5 parameters.
		if len(params) == 4 {
			params = append(params, -1)
		}
		if len(params) != 5 {
			return DrawCommand{}, fmt.Errorf("line requires 4 or 5 parameters: x1 y1 x2 y2 [colorIndex]")
		}
	case "lineto":
		// Accepts 2 or 3 parameters.
		if len(params) == 2 {
			params = append(params, -1)
		}
		if len(params) != 3 {
			return DrawCommand{}, fmt.Errorf("lineto requires 2 or 3 parameters: x y [colorIndex]")
		}
	case "rect":
		// Accepts 4 or 5 parameters.
		if len(params) == 4 {
			params = append(params, -1)
		}
		if len(params) != 5 {
			return DrawCommand{}, fmt.Errorf("rect requires 4 or 5 parameters: x y width height [colorIndex]")
		}
	case "cls", "flip":
		// "cls" takes no parameters; "flip" takes 0 or 1.
		if cmd == "cls" && len(params) != 0 {
			return DrawCommand{}, fmt.Errorf("cls takes no parameters")
		}
		if cmd == "flip" && len(params) > 1 {
			return DrawCommand{}, fmt.Errorf("flip takes 0 or 1 parameter (buffer index)")
		}
	case "ink":
		// Expect exactly 1 parameter (0–7).
		if len(params) != 1 {
			return DrawCommand{}, fmt.Errorf("ink requires 1 parameter (0–7)")
		}
		if params[0] < 0 || params[0] > 7 {
			return DrawCommand{}, fmt.Errorf("ink parameter must be between 0 and 7")
		}
	case "paper":
		// Expect exactly 1 parameter (0–7).
		if len(params) != 1 {
			return DrawCommand{}, fmt.Errorf("paper requires 1 parameter (0–7)")
		}
		if params[0] < 0 || params[0] > 7 {
			return DrawCommand{}, fmt.Errorf("paper parameter must be between 0 and 7")
		}
	case "bright":
		// Expect exactly 1 parameter (0 or 1).
		if len(params) != 1 {
			return DrawCommand{}, fmt.Errorf("bright requires 1 parameter (0 or 1)")
		}
		if params[0] != 0 && params[0] != 1 {
			return DrawCommand{}, fmt.Errorf("bright parameter must be 0 or 1")
		}
	case "colour":
		// Expect exactly 3 parameters: ink (0–7), paper (0–7), bright (0 or 1).
		if len(params) != 3 {
			return DrawCommand{}, fmt.Errorf("colour requires 3 parameters: ink paper bright")
		}
		if params[0] < 0 || params[0] > 7 {
			return DrawCommand{}, fmt.Errorf("colour: ink must be between 0 and 7")
		}
		if params[1] < 0 || params[1] > 7 {
			return DrawCommand{}, fmt.Errorf("colour: paper must be between 0 and 7")
		}
		if params[2] != 0 && params[2] != 1 {
			return DrawCommand{}, fmt.Errorf("colour: bright must be 0 or 1")
		}
	default:
		return DrawCommand{}, fmt.Errorf("unknown command %q", cmd)
	}
	return DrawCommand{Cmd: cmd, Params: params}, nil
}

// startGUIEventsServer listens on the provided address for GUI event clients.
func startGUIEventsServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error starting GUI events server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("GUI events server listening on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting GUI event connection:", err)
			continue
		}
		guiEventConnsMu.Lock()
		guiEventConns = append(guiEventConns, conn)
		guiEventConnsMu.Unlock()
		fmt.Println("New GUI event client connected:", conn.RemoteAddr())
	}
}

// sendGUIEvent broadcasts a GUI event string to all connected GUI event clients.
func sendGUIEvent(event string) {
	guiEventConnsMu.Lock()
	defer guiEventConnsMu.Unlock()
	for i := 0; i < len(guiEventConns); i++ {
		_, err := fmt.Fprintln(guiEventConns[i], event)
		if err != nil {
			// On error, close and remove the connection.
			guiEventConns[i].Close()
			guiEventConns = append(guiEventConns[:i], guiEventConns[i+1:]...)
			i--
		}
	}
}
