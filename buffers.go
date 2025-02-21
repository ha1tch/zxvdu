package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"strconv"
)

// TextureEntry holds a texture created from pixel data.
type TextureEntry struct {
	texture rl.Texture2D
	width   int
	height  int
	inUse   bool
}

// Global texture array (256 slots)
var textures [256]TextureEntry

// createFlipBuffers creates render textures for flip buffers (no alpha).
func createFlipBuffers() {
	internalW := BaseWidth * graphicsMult
	internalH := BaseHeight * graphicsMult
	flipBuffersMu.Lock()
	defer flipBuffersMu.Unlock()
	flipBuffers = make([]rl.RenderTexture2D, numFlipBuffers)
	for i := 0; i < numFlipBuffers; i++ {
		flipBuffers[i] = rl.LoadRenderTexture(int32(internalW), int32(internalH))
		rl.BeginTextureMode(flipBuffers[i])
		rl.ClearBackground(palette[effectivePaperColor()])
		rl.EndTextureMode()
	}
	activeFlipBuffer = 0
}

// createLayerBuffers creates render textures for layer buffers (with alpha).
func createLayerBuffers() {
	internalW := BaseWidth * graphicsMult
	internalH := BaseHeight * graphicsMult
	layerBuffersMu.Lock()
	defer layerBuffersMu.Unlock()
	layerBuffers = make([]rl.RenderTexture2D, numLayerBuffers)
	for i := 0; i < numLayerBuffers; i++ {
		layerBuffers[i] = rl.LoadRenderTexture(int32(internalW), int32(internalH))
		rl.BeginTextureMode(layerBuffers[i])
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
		rl.EndTextureMode()
	}
	activeLayerBuffer = 0
}

// createOffscreenBuffers creates render textures for offscreen buffers.
func createOffscreenBuffers() {
	internalW := BaseWidth * graphicsMult
	internalH := BaseHeight * graphicsMult
	
	flipBuffersMu.Lock()
	offscreenFlipBuffers = make([]rl.RenderTexture2D, numFlipBuffers)
	for i := 0; i < numFlipBuffers; i++ {
		offscreenFlipBuffers[i] = rl.LoadRenderTexture(int32(internalW), int32(internalH))
		rl.BeginTextureMode(offscreenFlipBuffers[i])
		rl.ClearBackground(palette[effectivePaperColor()])
		rl.EndTextureMode()
	}
	flipBuffersMu.Unlock()
	
	layerBuffersMu.Lock()
	offscreenLayerBuffers = make([]rl.RenderTexture2D, numLayerBuffers)
	for i := 0; i < numLayerBuffers; i++ {
		offscreenLayerBuffers[i] = rl.LoadRenderTexture(int32(internalW), int32(internalH))
		rl.BeginTextureMode(offscreenLayerBuffers[i])
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
		rl.EndTextureMode()
	}
	layerBuffersMu.Unlock()
}

// getActiveBuffer returns the currently active buffer based on drawing mode and target.
func getActiveBuffer() rl.RenderTexture2D {
	if currentTarget == "onscreen" {
		if currentDrawingMode == "flip" {
			return flipBuffers[activeFlipBuffer]
		}
		return layerBuffers[activeLayerBuffer]
	}
	
	if currentDrawingMode == "flip" {
		return offscreenFlipBuffers[activeOffscreenFlip]
	}
	return offscreenLayerBuffers[activeOffscreenLayer]
}

// copyBuffer copies contents from one buffer to another.
func copyBuffer(src, dst rl.RenderTexture2D) {
	rl.BeginTextureMode(dst)
	srcRect := rl.Rectangle{
		X: 0,
		Y: 0,
		Width: float32(src.Texture.Width),
		Height: float32(src.Texture.Height),
	}
	dstRect := rl.Rectangle{
		X: 0,
		Y: 0,
		Width: float32(dst.Texture.Width),
		Height: float32(dst.Texture.Height),
	}
	rl.DrawTexturePro(src.Texture, srcRect, dstRect, rl.Vector2{}, 0, rl.White)
	rl.EndTextureMode()
}

// copyBufferFromOffscreen copies a buffer from offscreen to onscreen.
func copyBufferFromOffscreen(bufferType string, srcIndex, dstIndex int) error {
	if srcIndex < 0 || dstIndex < 0 || srcIndex >= numFlipBuffers || dstIndex >= numFlipBuffers {
		return fmt.Errorf("buffer index out of range")
	}

	switch bufferType {
	case "flip":
		flipBuffersMu.Lock()
		defer flipBuffersMu.Unlock()
		copyBuffer(offscreenFlipBuffers[srcIndex], flipBuffers[dstIndex])
		
	case "layer":
		layerBuffersMu.Lock()
		defer layerBuffersMu.Unlock()
		copyBuffer(offscreenLayerBuffers[srcIndex], layerBuffers[dstIndex])
		
	default:
		return fmt.Errorf("invalid buffer type")
	}
	
	return nil
}

// findFirstFreeTextureSlot returns the index of the first free texture slot.
func findFirstFreeTextureSlot() int {
	for i := 0; i < len(textures); i++ {
		if !textures[i].inUse {
			return i
		}
	}
	return -1
}

// createTextureFromPixelData converts a hex pixel string into a Raylib texture.
func createTextureFromPixelData(pixelData string, width, height int) (rl.Texture2D, error) {
	if len(pixelData) != width*height {
		return rl.Texture2D{}, fmt.Errorf("pixel data length (%d) does not match dimensions %dx%d", len(pixelData), width, height)
	}
	
	// Create an image data slice.
	imgData := make([]rl.Color, width*height)
	for i, ch := range pixelData {
		// Validate hex character.
		val, err := strconv.ParseInt(string(ch), 16, 64)
		if err != nil {
			return rl.Texture2D{}, fmt.Errorf("invalid hex digit %q", ch)
		}
		// Check if within palette bounds (0-14).
		if val < 0 || val > 15 {
			return rl.Texture2D{}, fmt.Errorf("hex value %d out of range", val)
		}
		idx := int(val)
		if idx >= len(palette) {
			idx = len(palette) - 1
		}
		imgData[i] = palette[idx]
	}

	// Create a Raylib image.
	image := rl.GenImageColor(width, height, rl.Black)
	// Write our pixel data into the image.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			imgDataIndex := y*width + x
			rl.ImageDrawPixel(image, int32(x), int32(y), imgData[imgDataIndex])
		}
	}
	tex := rl.LoadTextureFromImage(image)
	rl.UnloadImage(image)
	return tex, nil
}

// updateTextureFromPixelData updates an existing texture from pixel data.
func updateTextureFromPixelData(slot int, pixelData string, width, height int) error {
	tex, err := createTextureFromPixelData(pixelData, width, height)
	if err != nil {
		return err
	}
	// Unload the previous texture if it exists.
	if textures[slot].inUse {
		rl.UnloadTexture(textures[slot].texture)
	}
	textures[slot] = TextureEntry{
		texture: tex,
		width:   width,
		height:  height,
		inUse:   true,
	}
	return nil
}

// deleteTexture frees a texture from the given slot.
func deleteTexture(slot int) error {
	if slot < 0 || slot >= len(textures) {
		return fmt.Errorf("texture number out of bounds")
	}
	if !textures[slot].inUse {
		return fmt.Errorf("texture slot %d is not in use", slot)
	}
	rl.UnloadTexture(textures[slot].texture)
	textures[slot] = TextureEntry{}
	return nil
}

// cleanupTextures releases all active textures.
func cleanupTextures() {
	for i := 0; i < len(textures); i++ {
		if textures[i].inUse {
			rl.UnloadTexture(textures[i].texture)
			textures[i] = TextureEntry{}
		}
	}
}