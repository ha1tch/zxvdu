package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"fmt"
)

// handleCLS clears the current active buffer
func handleCLS() {
	var rt rl.RenderTexture2D
	if currentTarget == "onscreen" {
		if currentDrawingMode == "flip" {
			flipBuffersMu.Lock()
			rt = flipBuffers[activeFlipBuffer]
			flipBuffersMu.Unlock()
		} else {
			layerBuffersMu.Lock()
			rt = layerBuffers[activeLayerBuffer]
			layerBuffersMu.Unlock()
		}
	} else {
		if currentDrawingMode == "flip" {
			flipBuffersMu.Lock()
			rt = offscreenFlipBuffers[activeOffscreenFlip]
			flipBuffersMu.Unlock()
		} else {
			layerBuffersMu.Lock()
			rt = offscreenLayerBuffers[activeOffscreenLayer]
			layerBuffersMu.Unlock()
		}
	}

	rl.BeginTextureMode(rt)
	if currentDrawingMode == "flip" {
		rl.ClearBackground(palette[effectivePaperColor()])
	} else {
		rl.ClearBackground(rl.Color{R: 0, G: 0, B: 0, A: 0})
	}
	rl.EndTextureMode()
}

// handleGraphics handles the graphics resolution multiplier command
func handleGraphics(cmd DrawCommand) {
	if len(cmd.Params) == 1 && cmd.Params[0] >= 1 {
		graphicsMult = cmd.Params[0]
		// Recreate all buffers
		createFlipBuffers()
		createLayerBuffers()
		createOffscreenBuffers()
		
		// Reset buffer indices
		activeFlipBuffer = 0
		activeLayerBuffer = 0
		activeOffscreenFlip = 0
		activeOffscreenLayer = 0
		
		// Update window size
		internalW := BaseWidth * graphicsMult
		internalH := BaseHeight * graphicsMult
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

// handleTexCommand processes texture-related commands
func handleTexCommand(cmd DrawCommand) {
	switch cmd.Mode {
	case "add":
		handleTexAdd(cmd)
	case "set":
		handleTexSet(cmd)
	case "del":
		handleTexDelete(cmd)
	case "paint":
		handleTexPaint(cmd)
	}
}

// handleTexAdd processes texture add command
func handleTexAdd(cmd DrawCommand) {
	if len(cmd.Params) < 2 {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0021 : invalid texture parameters\n"))
		}
		return
	}
	slot := findFirstFreeTextureSlot()
	if slot == -1 {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0022 : no free texture slots\n"))
		}
		return
	}
	tex, err := createTextureFromPixelData(cmd.Str, cmd.Params[0], cmd.Params[1])
	if err != nil {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0023 : " + err.Error() + "\n"))
		}
		return
	}
	textures[slot] = TextureEntry{
		texture: tex,
		width:   cmd.Params[0],
		height:  cmd.Params[1],
		inUse:   true,
	}
	// Send the texture slot number back to the client.
	if cmd.Conn != nil {
		fmt.Fprintln(cmd.Conn, slot)
	}
}

// handleTexSet processes texture set command
func handleTexSet(cmd DrawCommand) {
	if len(cmd.Params) < 3 {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0024 : invalid texture parameters\n"))
		}
		return
	}
	err := updateTextureFromPixelData(cmd.Params[0], cmd.Str, cmd.Params[1], cmd.Params[2])
	if err != nil {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0025 : " + err.Error() + "\n"))
		}
		return
	}
}

// handleTexDelete processes texture delete command
func handleTexDelete(cmd DrawCommand) {
	if len(cmd.Params) < 1 {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0026 : texture number required\n"))
		}
		return
	}
	err := deleteTexture(cmd.Params[0])
	if err != nil {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0027 : " + err.Error() + "\n"))
		}
		return
	}
}

// handleTexPaint processes texture paint command
func handleTexPaint(cmd DrawCommand) {
	if len(cmd.Params) < 3 {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0028 : invalid texture paint parameters\n"))
		}
		return
	}
	if cmd.Params[2] < 0 || cmd.Params[2] >= len(textures) || !textures[cmd.Params[2]].inUse {
		if cmd.Conn != nil {
			cmd.Conn.Write([]byte("ERROR 0029 : invalid texture number\n"))
		}
		return
	}
	rt := getTargetBuffer()
	rl.BeginTextureMode(rt)
	destRect := rl.Rectangle{
		X: float32(cmd.Params[0]),
		Y: float32(cmd.Params[1]),
		Width: float32(textures[cmd.Params[2]].width),
		Height: float32(textures[cmd.Params[2]].height),
	}
	srcRect := rl.Rectangle{
		X: 0,
		Y: 0,
		Width: float32(textures[cmd.Params[2]].width),
		Height: float32(textures[cmd.Params[2]].height),
	}
	rl.DrawTexturePro(textures[cmd.Params[2]].texture, srcRect, destRect, rl.Vector2{}, 0, rl.White)
	rl.EndTextureMode()
}