package main

import (
	"flag"
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Base resolution constants
const (
	BaseWidth  = 256
	BaseHeight = 192
)

// Global default drawing state
var (
	defaultInk    int  = 0  // ZX Spectrum default ink (foreground): black
	defaultPaper  int  = 7  // ZX Spectrum default paper (background): white
	defaultBright bool = false
)

// Global palette: ZX Spectrum 15-color palette
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

// Global state
var (
	currentX, currentY    int    = 0, 0    // For lineto commands
	currentDrawingMode    string = "flip"   // Current drawing mode
	buffers              *BufferSystem      // Global buffer system
	graphicsMult         int    = 1        // Graphics resolution multiplier 
	zoomFactor           int    = 1        // Display zoom factor
)

func main() {
	// Parse command-line flags
	inkFlag := flag.Int("ink", 0, "Default ink (foreground) color (0–7)")
	paperFlag := flag.Int("paper", 7, "Default paper (background) color (0–7)")
	brightFlag := flag.Int("bright", 0, "Default brightness flag (0 or 1)")
	hostFlag := flag.String("host", "0.0.0.0", "Server host address to bind to")
	cmdPortFlag := flag.String("cmdport", "55550", "Port for drawing command server")
	eventPortFlag := flag.String("eventport", "55551", "Port for event server")
	graphicsFlag := flag.Int("graphics", 1, "Graphics resolution multiplier")
	zoomFlag := flag.Int("zoom", 1, "Display zoom factor")
	flag.Parse()

	// Apply command line settings
	if *inkFlag >= 0 && *inkFlag <= 7 {
		defaultInk = *inkFlag
	}
	if *paperFlag >= 0 && *paperFlag <= 7 {
		defaultPaper = *paperFlag
	}
	defaultBright = (*brightFlag == 1)
	
	// Apply graphics and zoom settings
	if *graphicsFlag > 0 {
		graphicsMult = *graphicsFlag
	}
	if *zoomFlag > 0 {
		zoomFactor = *zoomFlag
	}

	// Calculate initial dimensions
	internalW := BaseWidth * graphicsMult
	internalH := BaseHeight * graphicsMult
	windowW := internalW * zoomFactor
	windowH := internalH * zoomFactor

	// Initialize window and rendering
	rl.InitWindow(int32(windowW), int32(windowH), "zxvdu - a simple VDU / display server")
	rl.SetTargetFPS(60)

	// Create buffer system
	buffers = NewBufferSystem(8, int32(internalW), int32(internalH))

	// Start network servers
	go startDrawingCommandServer(fmt.Sprintf("%s:%s", *hostFlag, *cmdPortFlag))
	go startEventServer(fmt.Sprintf("%s:%s", *hostFlag, *eventPortFlag))

	// Main render loop
	for !rl.WindowShouldClose() {
		processCommands()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Get the visible buffers (always buffer 0)
		flip, layer := buffers.GetDisplayBuffers()

		// Source rectangle for buffer content
		srcRect := rl.Rectangle{
			X: 0,
			Y: 0,
			Width: float32(internalW),
			Height: -float32(internalH), // Flip vertically
		}

		// Destination rectangle for scaled display
		dstRect := rl.Rectangle{
			X: 0,
			Y: 0,
			Width: float32(windowW),
			Height: float32(windowH),
		}

		// Draw flip buffer 0 (visible background)
		rl.DrawTexturePro(
			(*flip).Texture,
			srcRect,
			dstRect,
			rl.Vector2{},
			0,
			rl.White,
		)

		// Draw layer buffer 0 (visible overlay)
		rl.DrawTexturePro(
			(*layer).Texture,
			srcRect,
			dstRect,
			rl.Vector2{},
			0,
			rl.White,
		)

		// Handle mouse events
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()
			scaledX := int(mousePos.X) / zoomFactor
			scaledY := int(mousePos.Y) / zoomFactor
			eventStr := fmt.Sprintf("mouse: %d,%d", scaledX, scaledY)
			sendEvent(eventStr)
		}

		rl.EndDrawing()
	}

	// Cleanup
	buffers.Cleanup()
	rl.CloseWindow()
}