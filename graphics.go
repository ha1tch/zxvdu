package main

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

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

// getTargetBuffer returns the correct buffer based on current drawing state
func getTargetBuffer() rl.RenderTexture2D {
	if currentTarget == "onscreen" {
		if currentDrawingMode == "flip" {
			flipBuffersMu.RLock()
			defer flipBuffersMu.RUnlock()
			return flipBuffers[activeFlipBuffer]
		} else {
			layerBuffersMu.RLock()
			defer layerBuffersMu.RUnlock()
			return layerBuffers[activeLayerBuffer]
		}
	} else {
		if currentDrawingMode == "flip" {
			flipBuffersMu.RLock()
			defer flipBuffersMu.RUnlock()
			return offscreenFlipBuffers[activeOffscreenFlip]
		} else {
			layerBuffersMu.RLock()
			defer layerBuffersMu.RUnlock()
			return offscreenLayerBuffers[activeOffscreenLayer]
		}
	}
}

// updateActiveBuffer draws a command immediately into the active buffer.
func updateActiveBuffer(cmd DrawCommand) {
	rt := getTargetBuffer()
	rl.BeginTextureMode(rt)
	defer rl.EndTextureMode()

	// In layer mode, if eraser mode is active, drawing commands produce fully transparent pixels.
	var cOverride rl.Color
	if currentDrawingMode == "layer" && eraserMode {
		cOverride = rl.Color{R: 0, G: 0, B: 0, A: 0}
	}

	switch cmd.Cmd {
	case "plot":
		handlePlot(cmd, cOverride)
	case "line":
		handleLine(cmd, cOverride)
	case "lineto":
		handleLineTo(cmd, cOverride)
	case "circle":
		handleCircle(cmd, cOverride)
	case "rect":
		handleRect(cmd, cOverride)
	case "triangle":
		handleTriangle(cmd, cOverride)
	}
}

func handlePlot(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 2 {
		cIndex := -1
		if len(cmd.Params) >= 3 {
			cIndex = cmd.Params[2]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		rl.DrawPixel(int32(cmd.Params[0]), int32(cmd.Params[1]), col)
	}
}

func handleLine(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 4 {
		cIndex := -1
		if len(cmd.Params) >= 5 {
			cIndex = cmd.Params[4]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		rl.DrawLine(int32(cmd.Params[0]), int32(cmd.Params[1]),
			int32(cmd.Params[2]), int32(cmd.Params[3]), col)
	}
}

func handleLineTo(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 2 {
		cIndex := -1
		if len(cmd.Params) >= 3 {
			cIndex = cmd.Params[2]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		rl.DrawLine(int32(currentX), int32(currentY),
			int32(cmd.Params[0]), int32(cmd.Params[1]), col)
		currentX, currentY = cmd.Params[0], cmd.Params[1]
	}
}

func handleCircle(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 3 {
		cIndex := -1
		if len(cmd.Params) >= 4 {
			cIndex = cmd.Params[3]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		if strings.ToUpper(cmd.Mode) == "S" {
			rl.DrawCircleLines(int32(cmd.Params[0]), int32(cmd.Params[1]), float32(cmd.Params[2]), col)
		} else {
			rl.DrawCircle(int32(cmd.Params[0]), int32(cmd.Params[1]), float32(cmd.Params[2]), col)
		}
	}
}

func handleRect(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 4 {
		cIndex := -1
		if len(cmd.Params) >= 5 {
			cIndex = cmd.Params[4]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		if strings.ToUpper(cmd.Mode) == "S" {
			rl.DrawRectangleLines(int32(cmd.Params[0]), int32(cmd.Params[1]),
				int32(cmd.Params[2]), int32(cmd.Params[3]), col)
		} else {
			rl.DrawRectangle(int32(cmd.Params[0]), int32(cmd.Params[1]),
				int32(cmd.Params[2]), int32(cmd.Params[3]), col)
		}
	}
}

func handleTriangle(cmd DrawCommand, cOverride rl.Color) {
	if len(cmd.Params) >= 6 {
		cIndex := -1
		if len(cmd.Params) >= 7 {
			cIndex = cmd.Params[6]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		}
		var col rl.Color
		if currentDrawingMode == "layer" && eraserMode {
			col = cOverride
		} else {
			col = palette[cIndex]
		}
		p1 := rl.Vector2{X: float32(cmd.Params[0]), Y: float32(cmd.Params[1])}
		p2 := rl.Vector2{X: float32(cmd.Params[2]), Y: float32(cmd.Params[3])}
		p3 := rl.Vector2{X: float32(cmd.Params[4]), Y: float32(cmd.Params[5])}
		if strings.ToUpper(cmd.Mode) == "S" {
			rl.DrawLineV(p1, p2, col)
			rl.DrawLineV(p2, p3, col)
			rl.DrawLineV(p3, p1, col)
		} else {
			rl.DrawTriangle(p1, p2, p3, col)
		}
	}
}
