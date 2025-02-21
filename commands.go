package main

import (
	"fmt"
	"strconv"
	"strings"
)

// DrawCommand represents a drawing or control instruction
type DrawCommand struct {
	Cmd    string   // Command name
	Params []int    // Numeric parameters
	Mode   string   // Mode flags ("S"/"F"/"T" for shapes, "flip"/"layer" for paint)
	Str    string   // String data (used for texture data)
}

// parseCommand converts a text line into a DrawCommand
func parseCommand(line string) (DrawCommand, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("empty command")
	}
	cmd := strings.ToLower(fields[0])
	var dc DrawCommand
	dc.Cmd = cmd
	dc.Mode = "F" // default mode is fill

	// Handle query commands
	if len(fields) > 0 && fields[len(fields)-1] == "?" {
		fields = fields[:len(fields)-1]
		return parseQueryCommand(fields)
	}

	// Handle texture commands
	if cmd == "tex" {
		return parseTextureCommand(fields)
	}

	// Handle painting mode selection
	if cmd == "paint" {
		return parsePaintCommand(fields)
	}

	// Handle regular commands
	return parseRegularCommand(cmd, fields)
}

func parseTextureCommand(fields []string) (DrawCommand, error) {
	if len(fields) < 2 {
		return DrawCommand{}, fmt.Errorf("invalid texture command")
	}

	dc := DrawCommand{
		Cmd:  "tex",
		Mode: strings.ToLower(fields[1]), // add/set/del/paint
	}

	switch dc.Mode {
	case "add", "set":
		// tex add|set pixeldata width height
		// tex set n pixeldata width height
		if len(fields) < 4 {
			return dc, fmt.Errorf("insufficient parameters for tex %s", dc.Mode)
		}
		startIdx := 2
		if dc.Mode == "set" {
			if len(fields) < 5 {
				return dc, fmt.Errorf("insufficient parameters for tex set")
			}
			n, err := strconv.Atoi(fields[2])
			if err != nil {
				return dc, fmt.Errorf("invalid texture number")
			}
			dc.Params = append(dc.Params, n)
			startIdx = 3
		}
		dc.Str = fields[startIdx] // pixel data
		width, err := strconv.Atoi(fields[startIdx+1])
		if err != nil {
			return dc, fmt.Errorf("invalid width")
		}
		height, err := strconv.Atoi(fields[startIdx+2])
		if err != nil {
			return dc, fmt.Errorf("invalid height")
		}
		dc.Params = append(dc.Params, width, height)

	case "del":
		// tex del n
		if len(fields) < 3 {
			return dc, fmt.Errorf("texture number required")
		}
		n, err := strconv.Atoi(fields[2])
		if err != nil {
			return dc, fmt.Errorf("invalid texture number")
		}
		dc.Params = append(dc.Params, n)

	case "paint":
		// tex paint x y n
		if len(fields) < 5 {
			return dc, fmt.Errorf("insufficient parameters for tex paint")
		}
		x, err := strconv.Atoi(fields[2])
		if err != nil {
			return dc, fmt.Errorf("invalid x coordinate")
		}
		y, err := strconv.Atoi(fields[3])
		if err != nil {
			return dc, fmt.Errorf("invalid y coordinate")
		}
		n, err := strconv.Atoi(fields[4])
		if err != nil {
			return dc, fmt.Errorf("invalid texture number")
		}
		dc.Params = append(dc.Params, x, y, n)

	default:
		return dc, fmt.Errorf("unknown texture command mode: %s", dc.Mode)
	}

	return dc, nil
}

func parsePaintCommand(fields []string) (DrawCommand, error) {
	if len(fields) == 2 {
		if strings.ToLower(fields[1]) == "flip" {
			return DrawCommand{Cmd: "paint", Mode: "flip"}, nil
		}
		if strings.ToLower(fields[1]) == "layer" {
			return DrawCommand{Cmd: "paint", Mode: "layer"}, nil
		}
		// If not flip/layer, treat as buffer number
		n, err := strconv.Atoi(fields[1])
		if err != nil {
			return DrawCommand{}, fmt.Errorf("paint parameter must be number, flip, or layer")
		}
		return DrawCommand{Cmd: "paint", Params: []int{n}}, nil
	}
	return DrawCommand{Cmd: "paint", Params: []int{0}}, nil // Default to buffer 0
}

func parseQueryCommand(fields []string) (DrawCommand, error) {
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("empty query command")
	}
	
	return DrawCommand{
		Cmd: fields[0],
		Mode: "query",
	}, nil
}

func parseRegularCommand(cmd string, fields []string) (DrawCommand, error) {
	convertToken := func(token string) (int, error) {
		if token == "_" {
			return -1, nil
		}
		return strconv.Atoi(token)
	}

	switch cmd {
	case "plot", "line", "lineto", "ink", "paper", "bright", "colour", "cls", "flip", "layer":
		params := []int{}
		for _, token := range fields[1:] {
			val, err := convertToken(token)
			if err != nil {
				return DrawCommand{}, fmt.Errorf("invalid parameter %q", token)
			}
			params = append(params, val)
		}
		return DrawCommand{Cmd: cmd, Params: params}, nil

	case "rect", "circle", "triangle":
		return parseShapeCommand(cmd, fields)

	default:
		return DrawCommand{}, fmt.Errorf("unknown command %q", cmd)
	}
}

func parseShapeCommand(cmd string, fields []string) (DrawCommand, error) {
	params := []int{}
	tokenCount := len(fields) - 1
	mode := "F"

	if tokenCount > 0 {
		lastToken := strings.ToUpper(fields[len(fields)-1])
		if lastToken == "S" || lastToken == "F" || lastToken == "T" {
			mode = lastToken
			tokenCount--
		} else if _, err := strconv.Atoi(fields[len(fields)-1]); err != nil {
			return DrawCommand{}, fmt.Errorf("%s mode must be S, F, or T", cmd)
		}
	}

	// Convert numeric parameters
	for i := 1; i <= tokenCount; i++ {
		val, err := strconv.Atoi(fields[i])
		if err != nil {
			return DrawCommand{}, fmt.Errorf("invalid parameter %q", fields[i])
		}
		params = append(params, val)
	}

	// Validate parameter counts and add default color if needed
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

	return DrawCommand{
		Cmd:    cmd,
		Params: params,
		Mode:   mode,
	}, nil
}

// processQuery handles query commands
func processQuery(cmd string) string {
	switch cmd {
	case "colour":
		return fmt.Sprintf("%d %d %d", defaultInk, defaultPaper, boolToInt(defaultBright))
	case "ink":
		return fmt.Sprintf("%d", defaultInk)
	case "paper":
		return fmt.Sprintf("%d", defaultPaper)
	case "bright":
		return fmt.Sprintf("%d", boolToInt(defaultBright))
	case "paint":
		return currentDrawingMode
	case "host":
		return "zxvdu v1.0"
	default:
		return "unknown query"
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}