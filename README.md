# ZXVDU - ZX Spectrum-style Display Server

A modern network-capable display server that emulates ZX Spectrum-style graphics capabilities. Built with Go and raylib, it provides a flexible command interface for remote drawing operations with support for multiple buffer systems, texture management, and advanced compositing features.

## Key Features

- **Multi-Buffer System**
  - Flip buffers (opaque background)
  - Layer buffers (transparent overlay)
  - Offscreen buffers for composition
  - 8 buffers of each type available

- **Drawing Capabilities**
  - Full ZX Spectrum 15-color palette
  - Hardware-accelerated primitive rendering
  - Texture support with 256 slots
  - Transparency and compositing

- **Network Interface**
  - Command server (port 55550)
  - Event notification system (port 55551)
  - Text-based command protocol
  - Query system for state information

- **Display Management**
  - Configurable internal resolution (256×192 base)
  - Dynamic zoom support with monitor bounds checking
  - Independent flip and layer buffer control
  - Offscreen composition support

## Quick Start

### Prerequisites
- Go 1.16 or later
- raylib-go library

### Installation

```bash
# Clone the repository
git clone https://github.com/ha1tch/zxvdu.git
cd zxvdu

# Build the project
./mk.sh
```

Current builds support:
- macOS (Intel and Apple Silicon)
- Windows (32-bit and 64-bit Intel)

Community contributions for Linux and Raspberry Pi builds are welcome.

### Basic Configuration

```bash
# Example: 2x resolution, 2x zoom, white on black
./zxvdu -graphics=2 -zoom=2 -ink=7 -paper=0 -bright=1
```

## Command Reference

### Buffer Management

#### Target Selection
```bash
paint_target onscreen   # Draw to visible buffers
paint_target offscreen  # Draw to hidden buffers
```

#### Buffer Selection
```bash
paint flip             # Select flip buffer mode
paint layer            # Select layer buffer mode
flip [0-7]            # Select specific flip buffer
layer [0-7]           # Select specific layer buffer
```

#### Buffer Operations
```bash
cls                    # Clear current buffer
paint_copy flip 0 0    # Copy offscreen flip 0 to onscreen flip 0
paint_copy layer 1 2   # Copy offscreen layer 1 to onscreen layer 2
```

### Drawing Commands

#### Basic Drawing
```bash
plot x y [color]                    # Single pixel
line x1 y1 x2 y2 [color]           # Line between points
lineto x y [color]                  # Line from current position
```

#### Shapes
```bash
# Circles (F=fill, S=stroke)
circle x y radius [color] [F|S]     

# Rectangles
rect x y width height [color] [F|S] 

# Triangles
triangle x1 y1 x2 y2 x3 y3 [color] [F|S]
```

#### Color Control
```bash
ink n                  # Set foreground color (0-7)
paper n                # Set background color (0-7)
bright 0|1            # Toggle brightness
colour ink paper bright # Set all at once
```

### Texture Management
```bash
tex add pixeldata [width height]     # Create texture
tex set n pixeldata [width height]   # Update texture
tex del n                            # Delete texture
tex paint x y n                      # Draw texture
```

#### Creating Textures

Textures in ZXVDU use a simple hex format where each pixel is represented by a single hexadecimal digit (0-F) corresponding to the ZX Spectrum palette. 

To simplify texture creation, we recommend using [zxtex](https://github.com/ha1tch/zxtex), our companion tool for converting images to the required hex format. You can find zxtex, along with detailed documentation and examples, at:

https://github.com/ha1tch/zxtex

Example workflow:

1. First, convert your image to hex format using zxtex:
```bash
zxtex sprite.png --raw > sprite.hex
```

2. The hex data can then be copied from sprite.hex and used in a ZXVDU tex command:
```
tex add 7777000077770000 4 4     # Example: creating a 4x4 checkerboard pattern
```

The zxtex tool supports PNG, GIF, and BMP inputs and automatically maps colors to the ZX Spectrum palette. It also supports transparency using the '.' character in the hex output.

### Display Control
```bash
graphics n            # Set internal resolution multiplier
zoom n               # Set display zoom factor
```

## Query System

Append ? to commands for state queries:
```bash
colour?              # Returns "ink paper bright"
graphics?            # Returns current multiplier
zoom?                # Returns current zoom
paint_target?        # Returns current target
paint?               # Returns current mode
```

## Advanced Usage

### Double Buffering
```bash
paint_target offscreen
paint flip
cls
# Draw your scene
paint_copy flip 0 0
```

### Layer Composition
```bash
paint_target onscreen
paint flip
# Draw background
paint layer
# Draw overlays
```

### Texture Pipeline
```bash
tex add "0123..." 16 16  # Creates texture, returns n
tex paint 100 100 n      # Draws at position
tex del n                # Cleanup
```

## Contributing

Contributions are welcome, particularly:
- Linux build support
- Raspberry Pi build support
- Documentation enhancements

## License

Licensed under Apache License 2.0. See LICENSE file for details.

## Contact

- **Author:** haitch
- **Email:** haitch@duck.com
- **Social Media:** [@haitchfive@oldbytes.space](https://oldbytes.space/@haitchfive)
