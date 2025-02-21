package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"fmt"
)

// handleCLS clears the current active buffer
func handleCLS() {
	flip, layer := buffers.GetTargetBuffers()
	if currentDrawingMode == "flip" {
		rl.BeginTextureMode(*flip)
		rl.ClearBackground(palette[effectivePaperColor()])
		rl.EndTextureMode()
	} else {
		rl.BeginTextureMode(*layer)
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
		rl.EndTextureMode()
	}
}

// handleGraphics handles the graphics resolution multiplier command
func handleGraphics(cmd DrawCommand) {
	if len(cmd.Params) == 1 && cmd.Params[0] >= 1 {
		graphicsMult = cmd.Params[0]
		
		// Calculate new dimensions
		internalW := BaseWidth * graphicsMult
		internalH := BaseHeight * graphicsMult
		
		// Create new buffer system with updated dimensions
		buffers = NewBufferSystem(8, int32(internalW), int32(internalH))
		
		// Update window size
		rl.SetWindowSize(internalW*zoomFactor, internalH*zoomFactor)
	}
}

// handleZoom handles the zoom factor command
func handleZoom(cmd DrawCommand) {
	if len(cmd.Params) == 1 && cmd.Params[0] >= 1 {
		newZoom := cmd.Params[0]
		internalW := BaseWidth * graphicsMult
		internalH := BaseHeight * graphicsMult
		monW := rl.GetMonitorWidth(0)
		monH := rl.GetMonitorHeight(0)
		
		// Only change zoom if it fits within monitor bounds
		if internalW*newZoom <= monW && internalH*newZoom <= monH {
			zoomFactor = newZoom
			rl.SetWindowSize(internalW*zoomFactor, internalH*zoomFactor)
		}
	}
}

// handleTextureCapture processes a texture capture from screen region
func handleTextureCapture(cmd DrawCommand) (int, error) {
	if len(cmd.Params) < 4 {
		return -1, fmt.Errorf("invalid capture parameters")
	}

	// Validate region parameters
	if cmd.Params[0] < 0 || cmd.Params[1] < 0 || cmd.Params[2] <= 0 || cmd.Params[3] <= 0 {
		return -1, fmt.Errorf("invalid region bounds")
	}

	region := CaptureRegion{
		X:      cmd.Params[0],
		Y:      cmd.Params[1],
		Width:  cmd.Params[2],
		Height: cmd.Params[3],
	}

	// Get the appropriate source buffer based on current mode
	flip, layer := buffers.GetTargetBuffers()
	source := flip
	if currentDrawingMode == "layer" {
		source = layer
	}

	// Attempt to create texture from buffer region
	slot, err := CreateTextureFromBuffer(source, region)
	if err != nil {
		return -1, fmt.Errorf("texture capture failed: %v", err)
	}

	return slot, nil
}

// handleTexCommand processes texture-related commands
func handleTexCommand(cmd DrawCommand) (int, error) {
	switch cmd.Mode {
	case "add":
		if len(cmd.Params) < 2 {
			return -1, fmt.Errorf("invalid texture parameters")
		}
		if cmd.Str == "" {
			return -1, fmt.Errorf("no pixel data provided")
		}
		return CreateTextureFromPixelData(cmd.Str, cmd.Params[0], cmd.Params[1])

	case "set":
		if len(cmd.Params) < 3 {
			return -1, fmt.Errorf("invalid texture parameters")
		}
		if cmd.Str == "" {
			return -1, fmt.Errorf("no pixel data provided")
		}
		if cmd.Params[0] < 0 || cmd.Params[0] >= len(textures) || !textures[cmd.Params[0]].inUse {
			return -1, fmt.Errorf("invalid texture number")
		}
		// Delete existing texture
		rl.UnloadTexture(textures[cmd.Params[0]].texture)
        // Create new texture
		slot, err := CreateTextureFromPixelData(cmd.Str, cmd.Params[1], cmd.Params[2])
		if err != nil {
			textures[cmd.Params[0]] = TextureEntry{}
			return -1, err
		}
		return slot, nil

	case "del":
		if len(cmd.Params) < 1 {
			return -1, fmt.Errorf("texture number required")
		}
		if cmd.Params[0] < 0 || cmd.Params[0] >= len(textures) || !textures[cmd.Params[0]].inUse {
			return -1, fmt.Errorf("invalid texture number")
		}
		rl.UnloadTexture(textures[cmd.Params[0]].texture)
		textures[cmd.Params[0]] = TextureEntry{}
		return cmd.Params[0], nil

	case "paint":
		if len(cmd.Params) < 3 {
			return -1, fmt.Errorf("invalid texture paint parameters")
		}
		if cmd.Params[2] < 0 || cmd.Params[2] >= len(textures) || !textures[cmd.Params[2]].inUse {
			return -1, fmt.Errorf("invalid texture number")
		}
		flip, layer := buffers.GetTargetBuffers()
		target := flip
		if currentDrawingMode == "layer" {
			target = layer
		}
		rl.BeginTextureMode(*target)
		destRect := rl.Rectangle{
			X:      float32(cmd.Params[0]),
			Y:      float32(cmd.Params[1]),
			Width:  float32(textures[cmd.Params[2]].width),
			Height: float32(textures[cmd.Params[2]].height),
		}
		srcRect := rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(textures[cmd.Params[2]].width),
			Height: float32(textures[cmd.Params[2]].height),
		}
		rl.DrawTexturePro(textures[cmd.Params[2]].texture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
		rl.EndTextureMode()
		return cmd.Params[2], nil
	}

	return -1, fmt.Errorf("unknown texture command mode")
}