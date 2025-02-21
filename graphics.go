package main

import (
	"strings"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// effectiveInkColor computes the actual ink colour index (taking brightness into account)
func effectiveInkColor() int {
	if defaultInk == 0 {
		return 0
	}
	if defaultBright {
		return defaultInk + 7
	}
	return defaultInk
}

// effectivePaperColor computes the paper colour index (taking brightness into account)
func effectivePaperColor() int {
	if defaultPaper == 0 {
		return 0
	}
	if defaultBright {
		return defaultPaper + 7
	}
	return defaultPaper
}

// updateActiveBuffer draws a command immediately into the active buffer
func updateActiveBuffer(bs *BufferSystem, cmd DrawCommand, isLayer bool) (int, error) {
	flip, layer := bs.GetTargetBuffers()
	target := flip
	if isLayer {
		target = layer
	}

	rl.BeginTextureMode(*target)
	defer rl.EndTextureMode()

	var slot int
	var err error

	switch cmd.Cmd {
	case "plot":
		handlePlot(cmd)
	case "line":
		handleLine(cmd)
	case "lineto":
		handleLineTo(cmd)
	case "circle":
		handleCircle(cmd)
	case "rect":
		slot, err = handleRect(cmd, target)
	case "triangle":
		handleTriangle(cmd)
	}

	return slot, err
}

func handlePlot(cmd DrawCommand) {
	if len(cmd.Params) >= 2 {
		cIndex := -1
		if len(cmd.Params) >= 3 {
			cIndex = cmd.Params[2]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		} else if cIndex >= len(palette) {
			cIndex = len(palette) - 1
		}
		rl.DrawPixel(int32(cmd.Params[0]), int32(cmd.Params[1]), palette[cIndex])
	}
}

func handleLine(cmd DrawCommand) {
	if len(cmd.Params) >= 4 {
		cIndex := -1
		if len(cmd.Params) >= 5 {
			cIndex = cmd.Params[4]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		} else if cIndex >= len(palette) {
			cIndex = len(palette) - 1
		}
		rl.DrawLine(
			int32(cmd.Params[0]), int32(cmd.Params[1]),
			int32(cmd.Params[2]), int32(cmd.Params[3]),
			palette[cIndex],
		)
	}
}

func handleLineTo(cmd DrawCommand) {
	if len(cmd.Params) >= 2 {
		cIndex := -1
		if len(cmd.Params) >= 3 {
			cIndex = cmd.Params[2]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		} else if cIndex >= len(palette) {
			cIndex = len(palette) - 1
		}
		rl.DrawLine(
			int32(currentX), int32(currentY),
			int32(cmd.Params[0]), int32(cmd.Params[1]),
			palette[cIndex],
		)
		currentX, currentY = cmd.Params[0], cmd.Params[1]
	}
}

func handleCircle(cmd DrawCommand) {
	if len(cmd.Params) >= 3 {
		cIndex := -1
		if len(cmd.Params) >= 4 {
			cIndex = cmd.Params[3]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		} else if cIndex >= len(palette) {
			cIndex = len(palette) - 1
		}
		if strings.EqualFold(cmd.Mode, "S") {
			rl.DrawCircleLines(
				int32(cmd.Params[0]), int32(cmd.Params[1]),
				float32(cmd.Params[2]),
				palette[cIndex],
			)
		} else {
			rl.DrawCircle(
				int32(cmd.Params[0]), int32(cmd.Params[1]),
				float32(cmd.Params[2]),
				palette[cIndex],
			)
		}
	}
}

func handleRect(cmd DrawCommand, target *rl.RenderTexture2D) (int, error) {
	if len(cmd.Params) < 4 {
		return -1, nil
	}

	// Handle texture capture mode
	if strings.EqualFold(cmd.Mode, "T") {
		region := CaptureRegion{
			X:      cmd.Params[0],
			Y:      cmd.Params[1],
			Width:  cmd.Params[2],
			Height: cmd.Params[3],
		}
		return CreateTextureFromBuffer(target, region)
	}

	// Normal rectangle drawing
	cIndex := -1
	if len(cmd.Params) >= 5 {
		cIndex = cmd.Params[4]
	}
	if cIndex == -1 {
		cIndex = effectiveInkColor()
	} else if cIndex >= len(palette) {
		cIndex = len(palette) - 1
	}

	if strings.EqualFold(cmd.Mode, "S") {
		rl.DrawRectangleLines(
			int32(cmd.Params[0]), int32(cmd.Params[1]),
			int32(cmd.Params[2]), int32(cmd.Params[3]),
			palette[cIndex],
		)
	} else {
		rl.DrawRectangle(
			int32(cmd.Params[0]), int32(cmd.Params[1]),
			int32(cmd.Params[2]), int32(cmd.Params[3]),
			palette[cIndex],
		)
	}
	return -1, nil
}

func handleTriangle(cmd DrawCommand) {
	if len(cmd.Params) >= 6 {
		cIndex := -1
		if len(cmd.Params) >= 7 {
			cIndex = cmd.Params[6]
		}
		if cIndex == -1 {
			cIndex = effectiveInkColor()
		} else if cIndex >= len(palette) {
			cIndex = len(palette) - 1
		}
		p1 := rl.Vector2{X: float32(cmd.Params[0]), Y: float32(cmd.Params[1])}
		p2 := rl.Vector2{X: float32(cmd.Params[2]), Y: float32(cmd.Params[3])}
		p3 := rl.Vector2{X: float32(cmd.Params[4]), Y: float32(cmd.Params[5])}
		
		if strings.EqualFold(cmd.Mode, "S") {
			rl.DrawLineV(p1, p2, palette[cIndex])
			rl.DrawLineV(p2, p3, palette[cIndex])
			rl.DrawLineV(p3, p1, palette[cIndex])
		} else {
			rl.DrawTriangle(p1, p2, p3, palette[cIndex])
		}
	}
}