# zxvdu - a simple VDU / display server

A simple VDU (display) server built in Go using the [raylib-go](https://github.com/gen2brain/raylib-go) library. This project simulates a classic ZX Spectrum–style display with a 15‑colour palette and provides a network interface for remote drawing commands. It is designed to work with slow remote clients (such as a ZX Spectrum sending serial commands) by buffering and scaling the output appropriately.

---

## Features

- **Drawing Primitives:**  
  Supports a variety of commands to render:
  - **Pixel:** `plot x y [colorIndex]`
  - **Line:** `line x1 y1 x2 y2 [colorIndex]` and `lineto x y [colorIndex]`
  - **Circle:** `circle x y radius [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)
  - **Rectangle:** `rect x y width height [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)
  - **Triangle:** `triangle x1 y1 x2 y2 x3 y3 [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)

- **Colour Management:**  
  Set the drawing state with:
  - **ink:** `ink color` – Sets the default ink (foreground) colour (0–7).  
  - **paper:** `paper color` – Sets the default paper (background) colour (0–7).  
  - **bright:** `bright 0|1` – Sets the brightness flag (0 for off, 1 for on).  
  - **colour:** `colour ink paper bright` – Sets ink, paper, and brightness all at once.

  For drawing primitives, if the color parameter is omitted or specified as an underscore (`_`), the server uses the effective default ink colour (or paper colour for background operations).

- **Buffer Management:**  
  Manage drawing buffers with:
  - **cls:** Clears the current active buffer. When clearing, the background is painted using the effective paper colour.
  - **flip:** Toggles between or selects one of up to 8 buffers.  
    Usage: `flip` (toggle between buffers 0 and 1) or `flip bufferIndex`.

- **Resolution and Scaling:**  
  Two separate mechanisms manage resolution:
  - **Graphics Resolution:**  
    The internal (buffer) resolution is based on a base resolution of **256×192** multiplied by a factor.  
    Use the `graphics multiplier` command (or the command‑line flag) to set this multiplier. For example,  
    - `graphics 1` sets the internal resolution to 256×192 (default).  
    - `graphics 2` sets it to 512×384, and so on.  
    Changing the graphics multiplier resets all buffers and recreates the internal render target.
    
  - **Zoom Factor:**  
    The `zoom zoomFactor` command (or command‑line flag) scales the rendered internal buffer when displaying it on the host window. This affects only the display scaling (the internal buffer remains at its resolution).  
    The zoom factor can be changed at runtime without erasing the buffers. The server will adjust the window size accordingly if the resulting size fits within the host monitor’s dimensions.

- **Command‑Line Configuration:**  
  You can also set initial values for the following parameters when launching the application:
  - **ink** (`-ink`)
  - **paper** (`-paper`)
  - **bright** (`-bright`)
  - **graphics** (`-graphics`)
  - **zoom** (`-zoom`)

- **GUI Event Notifications:**  
  Clients connecting on a separate TCP port can receive GUI events (e.g., mouse clicks). The server broadcasts events such as `mouse: x,y`.

- **Network Interface:**  
  - **Drawing Commands Server:** Listens on TCP port **55550**.
  - **GUI Events Server:** Listens on TCP port **55551**.

---

## Getting Started

### Prerequisites

- [Go](https://golang.org/) (version 1.16 or later recommended)
- [raylib-go](https://github.com/gen2brain/raylib-go) installed and properly configured

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/yourusername/vdu-display-server.git
   cd vdu-display-server
   ```

2. **Install dependencies:**

   Ensure that `raylib-go` is installed:

   ```bash
   go get github.com/gen2brain/raylib-go/raylib
   ```

3. **Build the project:**

   ```bash
   go build -o vdu-display-server
   ```

4. **Run the server:**

   ```bash
   ./vdu-display-server
   ```

   You can also supply command‑line flags to set initial values:

   ```bash
   ./vdu-display-server -ink=3 -paper=0 -bright=1 -graphics=2 -zoom=2
   ```

   This example sets the default ink colour to 3, paper to 0, brightness on, internal resolution to 512×384 (2× multiplier), and a zoom factor of 2 (resulting in a window size of 1024×768).

---

## Usage

### Command Protocol

Commands are sent as single lines of text (terminated by a newline) with space‑separated parameters. Here are some examples:

#### Drawing Commands

- **Plot a Pixel:**

  ```bash
  plot 50 50 _
  ```

- **Draw a Line:**

  ```bash
  line 10 10 100 100 _
  ```

- **Draw a Circle:**

  - Filled (default):

    ```bash
    circle 200 200 50 _
    ```

  - Stroked:

    ```bash
    circle 200 200 50 _ S
    ```

- **Draw a Rectangle:**

  - Filled:

    ```bash
    rect 300 300 150 100 _
    ```

  - Stroked:

    ```bash
    rect 300 300 150 100 _ S
    ```

- **Draw a Triangle:**

  - Filled:

    ```bash
    triangle 100 100 150 50 200 100 _
    ```

  - Stroked:

    ```bash
    triangle 100 100 150 50 200 100 _ S
    ```

#### Colour and Drawing State

- **Set Default Ink:**

  ```bash
  ink 3
  ```

- **Set Default Paper:**

  ```bash
  paper 7
  ```

- **Set Brightness:**

  ```bash
  bright 1
  ```

- **Set All Colour Values:**

  ```bash
  colour 0 7 1
  ```

#### Buffer Commands

- **Clear Current Buffer:**

  ```bash
  cls
  ```

- **Flip Buffers:**

  Toggle (default between buffers 0 and 1):

  ```bash
  flip
  ```

  Or select a specific buffer:

  ```bash
  flip 3
  ```

#### Resolution and Scaling

- **Set Graphics Resolution:**

  This command resets all buffers and sets the internal resolution to base (256×192) multiplied by the given factor.

  ```bash
  graphics 1   # 256 x 192 (default)
  graphics 2   # 512 x 384
  ```

- **Set Zoom Factor:**

  Adjusts the display scaling (without clearing buffers):

  ```bash
  zoom 1   # 1:1 display (default)
  zoom 2   # Scales internal buffer by a factor of 2 (e.g., 512 x 384 becomes 1024 x 768 on screen)
  ```

#### GUI Event Notifications

Clients can connect on TCP port **55551** to receive events such as:

```bash
mouse: 123,456
```

---

## Network Interface

- **Drawing Commands:**  
  Connect to TCP port **55550** to send drawing and control commands.

- **GUI Events:**  
  Connect to TCP port **55551** to receive GUI event notifications.

---

## Project Structure

- **main.go:**  
  Contains the full implementation of the VDU/Display Server including:
  - Command parsing and handling.
  - Drawing primitives rendering.
  - Buffer management.
  - Graphics resolution and zoom management.
  - Network servers for drawing commands and GUI events.
  - Command‑line flag parsing for initial settings.

---

## Contact

**Name:** haitch  
**Email:** [haitch@duck.com](mailto:haitch@duck.com)  
**Social Media:** [https://oldbytes.space/@haitchfive](https://oldbytes.space/@haitchfive)


