package main

import (
	"flag"
	"fmt"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
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

// Global default drawing state.
var (
	defaultInk    int  = 7  // ZX Spectrum default ink (foreground): white.
	defaultPaper  int  = 0  // ZX Spectrum default paper (background): black.
	defaultBright bool = false
)

// Global palette: ZX Spectrum 15-colour palette.
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

// Buffer management state.
var (
	numFlipBuffers   = 8
	numLayerBuffers  = 8
	
	// Onscreen buffers
	flipBuffers      []rl.RenderTexture2D
	layerBuffers     []rl.RenderTexture2D
	activeFlipBuffer int = 0
	activeLayerBuffer int = 0
	
	// Offscreen buffers
	offscreenFlipBuffers  []rl.RenderTexture2D
	offscreenLayerBuffers []rl.RenderTexture2D
	activeOffscreenFlip   int = 0
	activeOffscreenLayer  int = 0

	// Buffer mutexes
	flipBuffersMu    sync.RWMutex
	layerBuffersMu   sync.RWMutex
)

// Drawing state
var (
	currentDrawingMode string = "flip"  // "flip" or "layer"
	currentTarget     string = "onscreen" // "onscreen" or "offscreen"
	eraserMode        bool   = false    // Only valid in layer draw mode
	currentX, currentY int    = 0, 0    // For lineto commands
)

func main() {
	// Parse command-line flags.
	inkFlag := flag.Int("ink", -1, "Default ink (foreground) colour (0–7).")
	paperFlag := flag.Int("paper", -1, "Default paper (background) colour (0–7).")
	brightFlag := flag.Int("bright", -1, "Default brightness flag (0 or 1).")
	graphicsFlag := flag.Int("graphics", -1, "Internal resolution multiplier (>=1).")
	zoomFlag := flag.Int("zoom", -1, "Display zoom factor (>=1).")
	hostFlag := flag.String("host", "0.0.0.0", "Server host address to bind to")
	cmdPortFlag := flag.String("cmdport", "55550", "Port for drawing command server")
	eventPortFlag := flag.String("eventport", "55551", "Port for event server")
	flag.Parse()

	// Apply command line settings
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

	// Start network servers
	go startDrawingCommandServer(fmt.Sprintf("%s:%s", *hostFlag, *cmdPortFlag))
	go startEventServer(fmt.Sprintf("%s:%s", *hostFlag, *eventPortFlag))

	// Calculate dimensions
	internalWidth := BaseWidth * graphicsMult
	internalHeight := BaseHeight * graphicsMult
	windowWidth := internalWidth * zoomFactor
	windowHeight := internalHeight * zoomFactor

	// Initialize window and buffers
	rl.InitWindow(int32(windowWidth), int32(windowHeight), "zxvdu - a simple VDU / display server")
	rl.SetTargetFPS(60)

	createFlipBuffers()
	createLayerBuffers()
	createOffscreenBuffers()

	// Main render loop
	for !rl.WindowShouldClose() {
		processCommands()

		// Render composite image
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Calculate rectangles for display
		destRect := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(internalWidth * zoomFactor),
			Height: float32(internalHeight * zoomFactor),
		}
		srcRect := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(internalWidth),
			Height: -float32(internalHeight),
		}

		// Draw flip buffer
		flipBuffersMu.RLock()
		rl.DrawTexturePro(flipBuffers[activeFlipBuffer].Texture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
		flipBuffersMu.RUnlock()

		// Draw layer buffer
		layerBuffersMu.RLock()
		rl.DrawTexturePro(layerBuffers[activeLayerBuffer].Texture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
		layerBuffersMu.RUnlock()

		// Handle mouse events
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()
			eventStr := fmt.Sprintf("mouse: %d,%d", int(mousePos.X), int(mousePos.Y))
			sendEvent(eventStr)
		}

		rl.EndDrawing()
	}

	// Cleanup
	cleanup()
	rl.CloseWindow()
}

// cleanup releases all resources
func cleanup() {
	flipBuffersMu.Lock()
	for _, tex := range flipBuffers {
		rl.UnloadRenderTexture(tex)
	}
	for _, tex := range offscreenFlipBuffers {
		rl.UnloadRenderTexture(tex)
	}
	flipBuffersMu.Unlock()

	layerBuffersMu.Lock()
	for _, tex := range layerBuffers {
		rl.UnloadRenderTexture(tex)
	}
	for _, tex := range offscreenLayerBuffers {
		rl.UnloadRenderTexture(tex)
	}
	layerBuffersMu.Unlock()

	cleanupTextures()
}