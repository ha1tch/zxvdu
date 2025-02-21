# ZXVDU Command Reference

## Buffer Management Commands

### Buffer Mode Selection
- `paint flip` - Switch to flip buffer mode for opaque drawing
- `paint layer` - Switch to layer buffer mode for transparent drawing

### Buffer Selection
- `paint N` - Select buffer number N (0-7)
- `flip N` - Swap flip buffer N with buffer 0
- `layer N` - Swap layer buffer N with buffer 0

### Buffer Operations
- `cls` - Clear current buffer
  - In flip mode: Clears to paper color
  - In layer mode: Clears to transparent

## Drawing Commands

### Basic Drawing
```
plot x y [color]              # Draw single pixel
line x1 y1 x2 y2 [color]     # Draw line between points
lineto x y [color]           # Draw line from current position to (x,y)
```
Parameters:
- x, y: Coordinates (0-255 at base resolution)
- color: Optional color index (0-7, or 8-14 if bright)
  - Defaults to current ink color if omitted

### Shapes
```
circle x y radius [color] [mode]        # Draw circle
rect x y width height [color] [mode]    # Draw rectangle
triangle x1 y1 x2 y2 x3 y3 [color] [mode] # Draw triangle
```
Parameters:
- x, y: Position coordinates
- radius: Circle radius
- width, height: Rectangle dimensions
- x1-x3, y1-y3: Triangle vertex coordinates
- color: Optional color index (0-7, or 8-14 if bright)
- mode: 
  - F (default) - Filled shape
  - S - Stroke (outline)
  - T (rect only) - Texture capture

## Color Commands

### Individual Settings
```
ink n      # Set foreground color (0-7)
paper n    # Set background color (0-7)
bright b   # Set brightness (0 or 1)
```

### Combined Setting
```
colour i p b  # Set ink, paper, and bright together
```
Parameters:
- i: Ink color (0-7)
- p: Paper color (0-7)
- b: Brightness (0 or 1)

## Texture Commands

### Texture Management
```
rect x y width height T    # Capture region as texture
tex add pixeldata w h      # Create texture from hex data
tex set n pixeldata w h    # Update existing texture
tex del n                  # Delete texture
tex paint x y n           # Draw texture
```
Parameters:
- x, y: Position coordinates
- width, height: Region dimensions
- n: Texture slot number (0-255)
- pixeldata: Hex string representing pixels
- w, h: Texture dimensions

Texture Data Format:
- One hex digit (0-F) per pixel
- Special characters:
  - . : Transparent pixel
  - @ : Light grey (7)
  - % : White (15)
  - ` : Black (0)

## Query Commands

Append ? to commands for state queries:
```
colour?        # Returns "ink paper bright"
ink?           # Returns current ink color
paper?         # Returns current paper color
bright?        # Returns current brightness
paint?         # Returns current mode (flip/layer)
host?          # Returns server version
```

## Server Configuration

Command-line flags when starting zxvdu:
```
-graphics N    # Resolution multiplier (default: 1)
-zoom N        # Display zoom factor (default: 1)
-ink N         # Initial ink color (default: 0)
-paper N       # Initial paper color (default: 7)
-bright N      # Initial brightness (default: 0)
-host addr     # Server address (default: 0.0.0.0)
-cmdport port  # Command port (default: 55550)
-eventport port # Event port (default: 55551)
```

## Error Responses

Error messages follow the format:
```
ERROR XXXX : description
```
Common error codes:
- 0020: Command parsing error
- 0021-0029: Texture operation errors
- 0030-0032: Region capture errors
- 0033: Server busy

## Network Protocol Notes

- Commands sent as text strings over TCP
- Each command terminated with newline
- Responses also newline-terminated
- Success response either empty or command-specific
- Event notifications sent on separate port (55551)
- Mouse events format: "mouse: x,y"

## Examples

Double-buffered drawing:
```
paint flip        # Select flip mode
paint 1           # Draw to buffer 1
cls               # Clear it
circle 128 96 40  # Draw something
flip 1            # Swap with buffer 0
```

Transparent overlay:
```
paint layer      # Select layer mode
circle 100 100 20 2 F  # Red filled circle
paint flip       # Back to flip mode
```

Texture capture and use:
```
rect 0 0 32 32 T     # Capture region to texture
                     # Returns texture slot N
tex paint 100 100 N  # Draw captured texture
```