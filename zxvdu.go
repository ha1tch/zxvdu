package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Global default drawing state.
var (
	defaultInk    int  = 7  // ZX Spectrum default ink (foreground): white.
	defaultPaper  int  = 0  // ZX Spectrum default paper (background): black.
	defaultBright bool = false
)

// Base resolution constants.
const (
	BaseWidth  = 256
	BaseHeight = 192
)

// Global graphics settings.
var (
	graphicsMult int = 1 // Multiplier for the internal buffer resolution (default 1 => 256x192)
	zoomFactor   int = 1 // Zoom factor for display (default 1: 1:1 mapping)
)

// Global palette: ZX Spectrum 15-colour palette.
// Indices 0-7 are the base colours; indices 8-14 are the "bright" variants.
var palette = []rl.Color{
	rl.Black,                      // 0: Black
	rl.NewColor(0, 0, 205, 255),     // 1: Blue
	rl.NewColor(205, 0, 0, 255),     // 2: Red
	rl.NewColor(205, 0, 205, 255),   // 3: Magenta
	rl.NewColor(0, 205, 0, 255),     // 4: Green
	rl.NewColor(0, 205, 205, 255),   // 5: Cyan
	rl.NewColor(205, 205, 0, 255),   // 6: Yellow
	rl.NewColor(205, 205, 205, 255), // 7: White (normal)
	rl.NewColor(0, 0, 255, 255),     // 8: Bright Blue
	rl.NewColor(255, 0, 0, 255),     // 9: Bright Red
	rl.NewColor(255, 0, 255, 255),   // 10: Bright Magenta
	rl.NewColor(0, 255, 0, 255),     // 11: Bright Green
	rl.NewColor(0, 255, 255, 255),   // 12: Bright Cyan
	rl.NewColor(255, 255, 0, 255),   // 13: Bright Yellow
	rl.NewColor(255, 255, 255, 255), // 14: Bright White
}

// We'll have up to 8 buffers. Each buffer is a render texture.
var (
	numBuffers   = 8
	buffers      []rl.RenderTexture2D
	activeBuffer int = 0
	buffersMu    sync.Mutex // Protects the buffers slice when recreating textures.
)

// Global render state for commands that need persistent state.
var (
	currentX, currentY int = 0, 0 // For lineto commands.
)

// DrawCommand represents a drawing or control instruction.
// For rect, circle, and triangle commands, Mode indicates stroke ("S") or fill ("F") – default is fill.
type DrawCommand struct {
	Cmd    string // Supported commands: plot, circle, line, lineto, rect, triangle, cls, flip, ink, paper, bright, colour, graphics, zoom
	Params []int  // Numeric parameters for commands.
	Mode   string // Optional mode for geometry commands ("S" for stroke, "F" for fill). Default "F".
}

// commandChan is used to pass commands from network connections to the main loop.
var commandChan = make(chan DrawCommand, 100)

// Global slice of active GUI event connections.
var (
	guiEventConns   = make([]net.Conn, 0)
	guiEventConnsMu sync.Mutex
)

func main() {
	// Parse command-line flags.
	inkFlag := flag.Int("ink", -1, "Default ink (foreground) colour (0–7).")
	paperFlag := flag.Int("paper", -1, "Default paper (background) colour (0–7).")
	brightFlag := flag.Int("bright", -1, "Default brightness flag (0 or 1).")
	graphicsFlag := flag.Int("graphics", -1, "Internal resolution multiplier (>=1).")
	zoomFlag := flag.Int("zoom", -1, "Display zoom factor (>=1).")
	flag.Parse()

	if *inkFlag != -1 {
		defaultInk = *inkFlag
	}
	if *paperFlag != -1 {
		defaultPaper = *paperFlag
	}
	if *brightFlag != -1 {
		defaultBright = (*brightFlag == 1)
	}
	if *graphicsFlag != -1 && *graphicsFlag >= 1 {
		graphicsMult = *graphicsFlag
	}
	if *zoomFlag != -1 && *zoomFlag >= 1 {
		zoomFactor = *zoomFlag
	}

	// Start TCP servers for drawing commands and GUI events.
	go startDrawingCommandServer(":55550")
	go startGUIEventsServer(":55551")

	// Calculate internal and window dimensions.
	internalWidth := BaseWidth * graphicsMult
	internalHeight := BaseHeight * graphicsMult
	windowWidth := internalWidth * zoomFactor
	windowHeight := internalHeight * zoomFactor

	// Initialize the window (must be done before creating textures).
	rl.InitWindow(int32(windowWidth), int32(windowHeight), "zxvdu - a simple VDU / display server")
	rl.SetTargetFPS(60)

	// Create the initial buffers (after the window is created).
	createBuffers()

	// Main loop.
	for !rl.WindowShouldClose() {
		processCommands()

		// Draw the active buffer's texture to the window scaled by the zoom factor.
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		internalWidth = BaseWidth * graphicsMult
		internalHeight = BaseHeight * graphicsMult
		destRect := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(internalWidth * zoomFactor),
			Height: float32(internalHeight * zoomFactor),
		}
		srcRect := rl.Rectangle{X: 0, Y: 0, Width: float32(internalWidth), Height: -float32(internalHeight)}
		buffersMu.Lock()
		rl.DrawTexturePro(buffers[activeBuffer].Texture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
		buffersMu.Unlock()

		// Check for GUI events and broadcast.
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()
			eventStr := fmt.Sprintf("mouse: %d,%d", int(mousePos.X), int(mousePos.Y))
			sendGUIEvent(eventStr)
		}
		rl.EndDrawing()
	}

	// Cleanup.
	buffersMu.Lock()
	for _, tex := range buffers {
		rl.UnloadRenderTexture(tex)
	}
	buffersMu.Unlock()
	rl.CloseWindow()
}

// createBuffers creates the render textures for each buffer.
func createBuffers() {
	internalW := BaseWidth * graphicsMult
	internalH := BaseHeight * graphicsMult
	buffersMu.Lock()
	defer buffersMu.Unlock()
	buffers = make([]rl.RenderTexture2D, numBuffers)
	for i := 0; i < numBuffers; i++ {
		buffers[i] = rl.LoadRenderTexture(int32(internalW), int32(internalH))
		// Initialize buffer with effective paper colour.
		rl.BeginTextureMode(buffers[i])
		rl.ClearBackground(palette[effectivePaperColor()])
		rl.EndTextureMode()
	}
	activeBuffer = 0
}

// updateActiveBuffer updates the active render texture with the given drawing command.
func updateActiveBuffer(cmd DrawCommand) {
	buffersMu.Lock()
	rt := buffers[activeBuffer]
	buffersMu.Unlock()

	rl.BeginTextureMode(rt)
	switch cmd.Cmd {
	case "plot":
		if len(cmd.Params) >= 2 {
			cIndex := -1
			if len(cmd.Params) >= 3 {
				cIndex = cmd.Params[2]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				rl.DrawPixel(int32(cmd.Params[0]), int32(cmd.Params[1]), palette[cIndex])
			}
		}
	case "line":
		if len(cmd.Params) >= 4 {
			cIndex := -1
			if len(cmd.Params) >= 5 {
				cIndex = cmd.Params[4]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				rl.DrawLine(int32(cmd.Params[0]), int32(cmd.Params[1]),
					int32(cmd.Params[2]), int32(cmd.Params[3]), palette[cIndex])
			}
		}
	case "lineto":
		if len(cmd.Params) >= 2 {
			cIndex := -1
			if len(cmd.Params) >= 3 {
				cIndex = cmd.Params[2]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				rl.DrawLine(int32(currentX), int32(currentY),
					int32(cmd.Params[0]), int32(cmd.Params[1]), palette[cIndex])
				currentX, currentY = cmd.Params[0], cmd.Params[1]
			}
		}
	case "circle":
		if len(cmd.Params) >= 3 {
			cIndex := -1
			if len(cmd.Params) >= 4 {
				cIndex = cmd.Params[3]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				if strings.ToUpper(cmd.Mode) == "S" {
					rl.DrawCircleLines(int32(cmd.Params[0]), int32(cmd.Params[1]), float32(cmd.Params[2]), palette[cIndex])
				} else {
					rl.DrawCircle(int32(cmd.Params[0]), int32(cmd.Params[1]), float32(cmd.Params[2]), palette[cIndex])
				}
			}
		}
	case "rect":
		if len(cmd.Params) >= 4 {
			cIndex := -1
			if len(cmd.Params) >= 5 {
				cIndex = cmd.Params[4]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				if strings.ToUpper(cmd.Mode) == "S" {
					rl.DrawRectangleLines(int32(cmd.Params[0]), int32(cmd.Params[1]),
						int32(cmd.Params[2]), int32(cmd.Params[3]), palette[cIndex])
				} else {
					rl.DrawRectangle(int32(cmd.Params[0]), int32(cmd.Params[1]),
						int32(cmd.Params[2]), int32(cmd.Params[3]), palette[cIndex])
				}
			}
		}
	case "triangle":
		if len(cmd.Params) >= 6 {
			cIndex := -1
			if len(cmd.Params) >= 7 {
				cIndex = cmd.Params[6]
			}
			if cIndex == -1 {
				cIndex = effectiveInkColor()
			}
			if cIndex < len(palette) {
				p1 := rl.Vector2{X: float32(cmd.Params[0]), Y: float32(cmd.Params[1])}
				p2 := rl.Vector2{X: float32(cmd.Params[2]), Y: float32(cmd.Params[3])}
				p3 := rl.Vector2{X: float32(cmd.Params[4]), Y: float32(cmd.Params[5])}
				if strings.ToUpper(cmd.Mode) == "S" {
					rl.DrawLineV(p1, p2, palette[cIndex])
					rl.DrawLineV(p2, p3, palette[cIndex])
					rl.DrawLineV(p3, p1, palette[cIndex])
				} else {
					rl.DrawTriangle(p1, p2, p3, palette[cIndex])
				}
			}
		}
	}
	rl.EndTextureMode()
}

// processCommands consumes all commands from the command channel and updates state or textures.
func processCommands() {
	for {
		select {
		case cmd := <-commandChan:
			switch cmd.Cmd {
			case "cls":
				buffersMu.Lock()
				rt := buffers[activeBuffer]
				buffersMu.Unlock()
				rl.BeginTextureMode(rt)
				rl.ClearBackground(palette[effectivePaperColor()])
				rl.EndTextureMode()
			case "flip":
				if len(cmd.Params) == 0 {
					if activeBuffer == 0 {
						activeBuffer = 1
					} else {
						activeBuffer = 0
					}
				} else if len(cmd.Params) == 1 {
					if cmd.Params[0] >= 0 && cmd.Params[0] < numBuffers {
						activeBuffer = cmd.Params[0]
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
			case "graphics":
				if len(cmd.Params) == 1 && cmd.Params[0] >= 1 {
					graphicsMult = cmd.Params[0]
					createBuffers()
					activeBuffer = 0
					internalW := BaseWidth * graphicsMult
					internalH := BaseHeight * graphicsMult
					rl.SetWindowSize(internalW*zoomFactor, internalH*zoomFactor)
				}
			case "zoom":
				if len(cmd.Params) == 1 && cmd.Params[0] >= 1 {
					newZoom := cmd.Params[0]
					internalW := BaseWidth * graphicsMult
					internalH := BaseHeight * graphicsMult
					monW := rl.GetMonitorWidth(0)
					monH := rl.GetMonitorHeight(0)
					if internalW*newZoom <= monW && internalH*newZoom <= monH {
						zoomFactor = newZoom
						rl.SetWindowSize(internalW*zoomFactor, internalH*zoomFactor)
					}
				}
			default:
				updateActiveBuffer(cmd)
			}
		default:
			return
		}
	}
}

// effectiveInkColor computes the actual ink colour index (taking brightness into account).
func effectiveInkColor() int {
	if defaultInk == 0 {
		return 0
	}
	if defaultBright {
		return defaultInk + 7
	}
	return defaultInk
}

// effectivePaperColor computes the paper colour index (taking brightness into account).
func effectivePaperColor() int {
	if defaultPaper == 0 {
		return 0
	}
	if defaultBright {
		return defaultPaper + 7
	}
	return defaultPaper
}

// startDrawingCommandServer listens on a TCP port for drawing commands.
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

// handleDrawingCommandConn reads commands from a TCP connection.
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
// It accepts "_" as a placeholder (converted to -1) for numeric parameters.
// For rect, circle, and triangle commands, an optional mode token ("S" or "F") may appear as the last token.
// The "graphics" and "zoom" commands expect one numeric parameter.
func parseCommand(line string) (DrawCommand, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("empty command")
	}
	cmd := strings.ToLower(fields[0])
	var dc DrawCommand
	dc.Cmd = cmd
	dc.Mode = "F" // default mode is fill

	// Helper: if token is "_" then return -1; otherwise convert to int.
	convertToken := func(token string) (int, error) {
		if token == "_" {
			return -1, nil
		}
		return strconv.Atoi(token)
	}

	switch cmd {
	case "plot", "line", "lineto", "ink", "paper", "bright", "colour", "cls", "flip":
		params := []int{}
		for _, token := range fields[1:] {
			val, err := convertToken(token)
			if err != nil {
				return DrawCommand{}, fmt.Errorf("invalid parameter %q", token)
			}
			params = append(params, val)
		}
		dc.Params = params
	case "rect", "circle", "triangle":
		params := []int{}
		tokenCount := len(fields) - 1
		if tokenCount > 0 {
			if _, err := strconv.Atoi(fields[len(fields)-1]); err != nil {
				modeCandidate := strings.ToUpper(fields[len(fields)-1])
				if modeCandidate != "S" && modeCandidate != "F" {
					return DrawCommand{}, fmt.Errorf("%s mode must be S or F", cmd)
				}
				dc.Mode = modeCandidate
				tokenCount--
			}
		}
		for i := 1; i <= tokenCount; i++ {
			val, err := convertToken(fields[i])
			if err != nil {
				return DrawCommand{}, fmt.Errorf("invalid parameter %q", fields[i])
			}
			params = append(params, val)
		}
		switch cmd {
		case "rect":
			if len(params) == 4 {
				params = append(params, -1)
			} else if len(params) != 5 {
				return DrawCommand{}, fmt.Errorf("rect requires 4 or 5 numeric parameters, plus optional mode")
			}
		case "circle":
			if len(params) == 3 {
				params = append(params, -1)
			} else if len(params) != 4 {
				return DrawCommand{}, fmt.Errorf("circle requires 3 or 4 numeric parameters, plus optional mode")
			}
		case "triangle":
			if len(params) == 6 {
				params = append(params, -1)
			} else if len(params) != 7 {
				return DrawCommand{}, fmt.Errorf("triangle requires 6 or 7 numeric parameters, plus optional mode")
			}
		}
		dc.Params = params
	case "graphics":
		params := []int{}
		for _, token := range fields[1:] {
			val, err := convertToken(token)
			if err != nil {
				return DrawCommand{}, fmt.Errorf("invalid parameter %q", token)
			}
			params = append(params, val)
		}
		if len(params) != 1 {
			return DrawCommand{}, fmt.Errorf("graphics requires exactly 1 numeric parameter: multiplier")
		}
		dc.Params = params
	case "zoom":
		params := []int{}
		for _, token := range fields[1:] {
			val, err := convertToken(token)
			if err != nil {
				return DrawCommand{}, fmt.Errorf("invalid parameter %q", token)
			}
			params = append(params, val)
		}
		if len(params) != 1 {
			return DrawCommand{}, fmt.Errorf("zoom requires exactly 1 numeric parameter: zoom factor")
		}
		dc.Params = params
	default:
		return DrawCommand{}, fmt.Errorf("unknown command %q", cmd)
	}
	return dc, nil
}

// startGUIEventsServer listens for GUI event connections on a TCP port.
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

// sendGUIEvent broadcasts a GUI event string to all connected clients.
func sendGUIEvent(event string) {
	guiEventConnsMu.Lock()
	defer guiEventConnsMu.Unlock()
	for i := 0; i < len(guiEventConns); i++ {
		_, err := fmt.Fprintln(guiEventConns[i], event)
		if err != nil {
			guiEventConns[i].Close()
			guiEventConns = append(guiEventConns[:i], guiEventConns[i+1:]...)
			i--
		}
	}
}
