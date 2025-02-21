package main

import (
	"fmt"
	"strconv"
	"strings"
)

// parseCommand converts a text line into a DrawCommand.
func parseCommand(line string) (DrawCommand, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("ERROR 0001 : empty command")
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

	// Special branch for texture commands
	if cmd == "tex" {
		return parseTexCommand(fields)
	}

	// Special branch for paint commands
	if cmd == "paint" || cmd == "paint_target" || cmd == "paint_copy" {
		return parsePaintCommand(fields)
	}

	// Special handling for "eraser"
	if cmd == "eraser" {
		return parseEraserCommand(fields)
	}

	// Handle regular commands
	return parseRegularCommand(cmd, fields)
}

func parseQueryCommand(fields []string) (DrawCommand, error) {
	if len(fields) == 0 {
		return DrawCommand{}, fmt.Errorf("ERROR 0001 : empty query command")
	}
	
	return DrawCommand{
		Cmd: fields[0],
		Mode: "query",
	}, nil
}

func parsePaintCommand(fields []string) (DrawCommand, error) {
	if len(fields) < 2 {
		return DrawCommand{}, fmt.Errorf("ERROR 0006 : paint command requires parameters")
	}

	cmd := strings.ToLower(fields[0])
	switch cmd {
	case "paint":
		modeParam := strings.ToLower(fields[1])
		if modeParam != "flip" && modeParam != "layer" {
			return DrawCommand{}, fmt.Errorf("ERROR 0006 : paint parameter must be flip or layer")
		}
		return DrawCommand{
			Cmd:  "paint",
			Mode: modeParam,
		}, nil

	case "paint_target":
		targetParam := strings.ToLower(fields[1])
		if targetParam != "onscreen" && targetParam != "offscreen" {
			return DrawCommand{}, fmt.Errorf("ERROR 0006 : paint_target parameter must be onscreen or offscreen")
		}
		return DrawCommand{
			Cmd:  "paint_target",
			Mode: targetParam,
		}, nil

	case "paint_copy":
		if len(fields) != 4 {
			return DrawCommand{}, fmt.Errorf("ERROR 0006 : paint_copy requires buffer_type src dst")
		}
		bufferType := strings.ToLower(fields[1])
		if bufferType != "flip" && bufferType != "layer" {
			return DrawCommand{}, fmt.Errorf("ERROR 0006 : paint_copy buffer_type must be flip or layer")
		}
		src, err1 := strconv.Atoi(fields[2])
		dst, err2 := strconv.Atoi(fields[3])
		if err1 != nil || err2 != nil {
			return DrawCommand{}, fmt.Errorf("ERROR 0006 : invalid buffer numbers")
		}
		return DrawCommand{
			Cmd:    "paint_copy",
			Mode:   bufferType,
			Params: []int{src, dst},
		}, nil
	}

	return DrawCommand{}, fmt.Errorf("ERROR 0006 : invalid paint command")
}

func parseTexCommand(fields []string) (DrawCommand, error) {
	if len(fields) < 2 {
		return DrawCommand{}, fmt.Errorf("ERROR 0010 : tex command requires a subcommand")
	}
	
	subCmd := strings.ToLower(fields[1])
	var dc DrawCommand
	dc.Cmd = "tex"
	
	switch subCmd {
	case "add":
		return parseTexAdd(fields)
	case "set":
		return parseTexSet(fields)
	case "del":
		return parseTexDelete(fields)
	case "paint":
		return parseTexPaint(fields)
	default:
		return DrawCommand{}, fmt.Errorf("ERROR 0020 : unknown tex sub-command %q", subCmd)
	}
}

func parseTexAdd(fields []string) (DrawCommand, error) {
	if len(fields) < 3 {
		return DrawCommand{}, fmt.Errorf("ERROR 0011 : tex add requires pixel data")
	}
	
	dc := DrawCommand{
		Cmd:  "tex",
		Mode: "add",
		Str:  fields[2],
	}
	
	if len(fields) >= 5 {
		sx, err1 := strconv.Atoi(fields[3])
		sy, err2 := strconv.Atoi(fields[4])
		if err1 != nil || err2 != nil {
			return DrawCommand{}, fmt.Errorf("ERROR 0012 : invalid size parameters")
		}
		dc.Params = []int{sx, sy}
	} else {
		dc.Params = []int{16, 16} // Default size
	}
	
	return dc, nil
}

func parseTexSet(fields []string) (DrawCommand, error) {
	if len(fields) < 4 {
		return DrawCommand{}, fmt.Errorf("ERROR 0013 : tex set requires texture number and pixel data")
	}
	
	num, err := strconv.Atoi(fields[2])
	if err != nil {
		return DrawCommand{}, fmt.Errorf("ERROR 0014 : invalid texture number")
	}
	
	dc := DrawCommand{
		Cmd:    "tex",
		Mode:   "set",
		Params: []int{num},
		Str:    fields[3],
	}
	
	if len(fields) >= 6 {
		sx, err1 := strconv.Atoi(fields[4])
		sy, err2 := strconv.Atoi(fields[5])
		if err1 != nil || err2 != nil {
			return DrawCommand{}, fmt.Errorf("ERROR 0015 : invalid size parameters")
		}
		dc.Params = append(dc.Params, sx, sy)
	} else {
		dc.Params = append(dc.Params, 16, 16) // Default size
	}
	
	return dc, nil
}

func parseTexDelete(fields []string) (DrawCommand, error) {
	if len(fields) < 3 {
		return DrawCommand{}, fmt.Errorf("ERROR 0016 : tex del requires texture number")
	}
	
	num, err := strconv.Atoi(fields[2])
	if err != nil {
		return DrawCommand{}, fmt.Errorf("ERROR 0017 : invalid texture number")
	}
	
	return DrawCommand{
		Cmd:    "tex",
		Mode:   "del",
		Params: []int{num},
	}, nil
}

func parseTexPaint(fields []string) (DrawCommand, error) {
	if len(fields) < 5 {
		return DrawCommand{}, fmt.Errorf("ERROR 0021 : tex paint requires x, y, and texture number")
	}
	
	x, err1 := strconv.Atoi(fields[2])
	y, err2 := strconv.Atoi(fields[3])
	texNum, err3 := strconv.Atoi(fields[4])
	if err1 != nil || err2 != nil || err3 != nil {
		return DrawCommand{}, fmt.Errorf("ERROR 0022 : invalid parameters for tex paint")
	}
	
	return DrawCommand{
		Cmd:    "tex",
		Mode:   "paint",
		Params: []int{x, y, texNum},
	}, nil
}

func parseEraserCommand(fields []string) (DrawCommand, error) {
	if len(fields) != 1 {
		return DrawCommand{}, fmt.Errorf("ERROR 0007 : eraser command takes no parameters")
	}
	return DrawCommand{Cmd: "eraser"}, nil
}

func parseRegularCommand(cmd string, fields []string) (DrawCommand, error) {
	convertToken := func(token string) (int, error) {
		if token == "_" {
			return -1, nil
		}
		return strconv.Atoi(token)
	}

	switch cmd {
	case "plot", "line", "lineto", "ink", "paper", "bright", "colour", "cls", "flip", "layer", "graphics", "zoom":
		params := []int{}
		for _, token := range fields[1:] {
			val, err := convertToken(token)
			if err != nil {
				return DrawCommand{}, fmt.Errorf("ERROR 0002 : invalid parameter %q", token)
			}
			params = append(params, val)
		}
		return DrawCommand{Cmd: cmd, Params: params}, nil

	case "rect", "circle", "triangle":
		return parseShapeCommand(cmd, fields)

	default:
		return DrawCommand{}, fmt.Errorf("ERROR 0005 : unknown command %q", cmd)
	}
}

func parseShapeCommand(cmd string, fields []string) (DrawCommand, error) {
	params := []int{}
	tokenCount := len(fields) - 1
	mode := "F"

	if tokenCount > 0 {
		if _, err := strconv.Atoi(fields[len(fields)-1]); err != nil {
			modeCandidate := strings.ToUpper(fields[len(fields)-1])
			if modeCandidate != "S" && modeCandidate != "F" {
				return DrawCommand{}, fmt.Errorf("ERROR 0003 : %s mode must be S or F", cmd)
			}
			mode = modeCandidate
			tokenCount--
		}
	}

	// Convert numeric parameters
	for i := 1; i <= tokenCount; i++ {
		val, err := strconv.Atoi(fields[i])
		if err != nil {
			return DrawCommand{}, fmt.Errorf("ERROR 0002 : invalid parameter %q", fields[i])
		}
		params = append(params, val)
	}

	// Validate parameter counts and add default color if needed
	switch cmd {
	case "rect":
		if len(params) == 4 {
			params = append(params, -1)
		} else if len(params) != 5 {
			return DrawCommand{}, fmt.Errorf("ERROR 0004 : rect requires 4 or 5 numeric parameters, plus optional mode")
		}
	case "circle":
		if len(params) == 3 {
			params = append(params, -1)
		} else if len(params) != 4 {
			return DrawCommand{}, fmt.Errorf("ERROR 0004 : circle requires 3 or 4 numeric parameters, plus optional mode")
		}
	case "triangle":
		if len(params) == 6 {
			params = append(params, -1)
		} else if len(params) != 7 {
			return DrawCommand{}, fmt.Errorf("ERROR 0004 : triangle requires 6 or 7 numeric parameters, plus optional mode")
		}
	}

	return DrawCommand{
		Cmd:    cmd,
		Params: params,
		Mode:   mode,
	}, nil
}