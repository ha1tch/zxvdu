package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
	"sync"
)

// TextureEntry holds a texture created from pixel data
type TextureEntry struct {
	texture rl.Texture2D
	width   int
	height  int
	inUse   bool
}

// BufferSystem manages the display buffers
type BufferSystem struct {
	flipBuffers  []*rl.RenderTexture2D
	layerBuffers []*rl.RenderTexture2D
	activeTarget int
	mu          sync.RWMutex
}

// Global texture array (256 slots)
var textures [256]TextureEntry

// CaptureRegion represents a rectangular region to capture
type CaptureRegion struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewBufferSystem creates a new buffer system with the specified number of buffers
func NewBufferSystem(numBuffers int, width, height int32) *BufferSystem {
	bs := &BufferSystem{
		flipBuffers:  make([]*rl.RenderTexture2D, numBuffers),
		layerBuffers: make([]*rl.RenderTexture2D, numBuffers),
		activeTarget: 0,
	}

	// Initialize all buffers
	for i := 0; i < numBuffers; i++ {
		// Create flip buffer
		rt := rl.LoadRenderTexture(width, height)
		bs.flipBuffers[i] = &rt
		
		// Initialize with paper color
		rl.BeginTextureMode(rt)
		rl.ClearBackground(palette[effectivePaperColor()])
		rl.EndTextureMode()

		// Create layer buffer
		rt = rl.LoadRenderTexture(width, height)
		bs.layerBuffers[i] = &rt
		
		// Initialize transparent
		rl.BeginTextureMode(rt)
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
		rl.EndTextureMode()
	}

	return bs
}

// GetDisplayBuffers returns buffer 0 of each type (always visible)
func (bs *BufferSystem) GetDisplayBuffers() (*rl.RenderTexture2D, *rl.RenderTexture2D) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.flipBuffers[0], bs.layerBuffers[0]
}

// GetTargetBuffers returns the current flip and layer buffers for drawing
func (bs *BufferSystem) GetTargetBuffers() (*rl.RenderTexture2D, *rl.RenderTexture2D) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.flipBuffers[bs.activeTarget], bs.layerBuffers[bs.activeTarget]
}

// SwapFlip swaps flip buffer n with buffer 0
func (bs *BufferSystem) SwapFlip(n int) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if n < 1 || n >= len(bs.flipBuffers) {
		return fmt.Errorf("invalid buffer index")
	}

	bs.flipBuffers[0], bs.flipBuffers[n] = bs.flipBuffers[n], bs.flipBuffers[0]
	return nil
}

// SwapLayer swaps layer buffer n with buffer 0
func (bs *BufferSystem) SwapLayer(n int) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if n < 1 || n >= len(bs.layerBuffers) {
		return fmt.Errorf("invalid buffer index")
	}

	bs.layerBuffers[0], bs.layerBuffers[n] = bs.layerBuffers[n], bs.layerBuffers[0]
	return nil
}

// SetActiveTarget sets which buffer pair to draw to
func (bs *BufferSystem) SetActiveTarget(n int) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if n < 0 || n >= len(bs.flipBuffers) {
		return fmt.Errorf("invalid target")
	}

	bs.activeTarget = n
	return nil
}

// ClearFlip clears the active flip buffer to paper color
func (bs *BufferSystem) ClearFlip() {
	flip, _ := bs.GetTargetBuffers()
	rl.BeginTextureMode(*flip)
	rl.ClearBackground(palette[effectivePaperColor()])
	rl.EndTextureMode()
}

// ClearLayer clears the active layer buffer to transparent
func (bs *BufferSystem) ClearLayer() {
	_, layer := bs.GetTargetBuffers()
	rl.BeginTextureMode(*layer)
	rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
	rl.EndTextureMode()
}

// CreateTextureFromBuffer creates a texture from a region of a buffer
func CreateTextureFromBuffer(source *rl.RenderTexture2D, region CaptureRegion) (int, error) {
	// Find a free texture slot
	slot := findFirstFreeTextureSlot()
	if slot == -1 {
		return -1, fmt.Errorf("no free texture slots")
	}

	// Validate region bounds
	if region.X < 0 || region.Y < 0 || region.Width <= 0 || region.Height <= 0 ||
		region.X+region.Width > int(source.Texture.Width) ||
		region.Y+region.Height > int(source.Texture.Height) {
		return -1, fmt.Errorf("invalid region bounds")
	}

	// Get pixel data from the region
	rl.BeginTextureMode(*source)
	img := rl.LoadImageFromTexture(source.Texture)
	rl.ImageCrop(img, rl.Rectangle{
		X:      float32(region.X),
		Y:      float32(region.Y),
		Width:  float32(region.Width),
		Height: float32(region.Height),
	})
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rl.EndTextureMode()

	// Store in texture system
	textures[slot] = TextureEntry{
		texture: tex,
		width:   region.Width,
		height:  region.Height,
		inUse:   true,
	}

	return slot, nil
}

// CreateTextureFromPixelData creates a texture from provided hex string data
func CreateTextureFromPixelData(pixelData string, width, height int) (int, error) {
	// Find a free texture slot
	slot := findFirstFreeTextureSlot()
	if slot == -1 {
		return -1, fmt.Errorf("no free texture slots")
	}

	// Validate data length
	if len(pixelData) != width*height {
		return -1, fmt.Errorf("pixel data length (%d) does not match dimensions %dx%d", len(pixelData), width, height)
	}

	// Create image data
	imgData := make([]rl.Color, width*height)
	for i, ch := range pixelData {
		var idx int
		switch ch {
		case '.':
			// Transparent pixel
			imgData[i] = rl.Color{R: 0, G: 0, B: 0, A: 0}
			continue
		case '@':
			// Light grey (palette index 7)
			idx = 7
		case '%':
			// White (palette index 15)
			idx = 15
		case '`':
			// Black (palette index 0)
			idx = 0
		default:
			// Try to parse as hex
			val, err := strconv.ParseInt(string(ch), 16, 64)
			if err != nil {
				return -1, fmt.Errorf("invalid character %q - must be hex digit or one of: . @ % `", ch)
			}
			if val < 0 || val > 15 {
				return -1, fmt.Errorf("hex value %d out of range", val)
			}
			idx = int(val)
		}
		if idx >= len(palette) {
			idx = len(palette) - 1
		}
		imgData[i] = palette[idx]
	}

	// Create image and texture
	image := rl.GenImageColor(width, height, rl.Black)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			imgDataIndex := y*width + x
			rl.ImageDrawPixel(image, int32(x), int32(y), imgData[imgDataIndex])
		}
	}
	tex := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)

	// Store in texture system
	textures[slot] = TextureEntry{
		texture: tex,
		width:   width,
		height:  height,
		inUse:   true,
	}

	return slot, nil
}

// findFirstFreeTextureSlot returns the index of the first free texture slot
func findFirstFreeTextureSlot() int {
	for i := 0; i < len(textures); i++ {
		if !textures[i].inUse {
			return i
		}
	}
	return -1
}

// Cleanup releases all buffer resources
func (bs *BufferSystem) Cleanup() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	for i := 0; i < len(bs.flipBuffers); i++ {
		if bs.flipBuffers[i] != nil {
			rl.UnloadRenderTexture(*bs.flipBuffers[i])
		}
		if bs.layerBuffers[i] != nil {
			rl.UnloadRenderTexture(*bs.layerBuffers[i])
		}
	}

	// Cleanup textures
	for i := 0; i < len(textures); i++ {
		if textures[i].inUse {
			rl.UnloadTexture(textures[i].texture)
			textures[i] = TextureEntry{}
		}
	}
}