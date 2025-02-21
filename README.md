# ZXVDU - ZX Spectrum-style Display Server

A modern network-capable display server that emulates ZX Spectrum-style graphics capabilities. Built with Go and raylib, it provides a flexible command interface for remote drawing operations with support for multiple buffer systems, texture management, and advanced compositing features.

**Note:** This is pre-release software (v0.2.0). See [CHANGELOG.md](CHANGELOG.md) for version history and breaking changes.

## Features

### Graphics System
- Full ZX Spectrum 15-color palette support
- Hardware-accelerated primitive rendering (lines, circles, rectangles, triangles)
- Support for both filled and stroked shapes
- Texture capture from screen regions using rect T command
- 256 texture slots for storing and reusing graphics
- Configurable resolution and zoom settings

### Buffer Management
- Flip buffers for opaque drawing (paint flip)
- Layer buffers for transparent overlays (paint layer)
- 8 buffers of each type
- Double buffering support for smooth animation
- Efficient buffer swapping operations

### Network Interface
- Command server (port 55550)
- Event notification system (port 55551)
- Text-based command protocol with standard error format
- State query system
- Mouse event reporting with proper scaling

## Installation

### Prerequisites
- Go 1.16 or later
- raylib-go library

### Building
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

## Quick Start

### Starting the Server
```bash
# Basic startup with default settings
./zxvdu

# With custom settings
./zxvdu -graphics=2 -zoom=2 -ink=7 -paper=0 -bright=1
```

### Basic Usage
```bash
# Connect to server (using netcat)
nc localhost 55550

# Draw a red circle
paint flip
ink 2
bright 1
circle 128 96 40 _ F

# Capture region as texture
rect 0 0 32 32 T    # Returns texture slot number
tex paint 100 100 0  # Draw texture from slot 0

# Query current settings
colour?
2 0 1
```

## Documentation

- [COMMANDS.md](COMMANDS.md) - Complete command reference and protocol details
- [CHANGELOG.md](CHANGELOG.md) - Version history and breaking changes

## Experimental Features

### Python Module (Pre-release)
A Python module for ZXVDU is available in [examples/lib/zxvdu.py](examples/lib/zxvdu.py). This module is experimental and its API may change significantly. It provides:
- High-level interface to ZXVDU
- Buffer management abstraction
- Texture handling
- Error management
- Resource cleanup

**Note:** The Python module is under active development and is not yet recommended for production use.

### Example Programs

The following examples demonstrate ZXVDU features using the experimental Python module:

#### Bouncing Ball ([examples/ball/](examples/ball/))
- Double buffered animation
- Easing functions
- Proper buffer management
- See [examples/ball/README.md](examples/ball/README.md) for details

#### Moving Truck ([examples/truck/](examples/truck/))
- Layer compositing
- Texture capture and reuse
- Animated sprites
- See [examples/truck/README.md](examples/truck/README.md) for details

#### Space Invaders ([examples/invaders/](examples/invaders/))
- Game development example
- Sprite animation
- Collision detection
- See [examples/invaders/README.md](examples/invaders/README.md) for details

**Note:** These examples are provided for learning purposes and to demonstrate ZXVDU capabilities. They are subject to change as the Python module evolves.

## Error Handling

Error messages follow the format:
```
ERROR XXXX : description
```

Common error codes:
- 0020: Command parsing error
- 0021-0029: Texture operation errors
- 0030-0032: Region capture errors
- 0033: Server busy

For a complete list of error codes, see [COMMANDS.md](COMMANDS.md).

## Contributing

Contributions are welcome! Particularly interested in:
- Linux build support
- Raspberry Pi build support
- Additional examples
- Documentation improvements
- Python module enhancements

## License

Licensed under Apache License 2.0. See LICENSE file for details.

## Contact

- **Author:** haitch
- **Email:** haitch@duck.com
- **Social Media:** [@haitchfive@oldbytes.space](https://oldbytes.space/@haitchfive)