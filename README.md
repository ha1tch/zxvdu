# zxvdu - a simple VDU / display server

A simple VDU (display) server built in Go using the [raylib-go](https://github.com/gen2brain/raylib-go) library. This project simulates a classic ZX Spectrum–style display with a 15‑colour palette and provides a network interface for remote drawing commands. It is designed to work with slow remote clients (such as a ZX Spectrum sending serial commands) by buffering and scaling the output appropriately.

---

## Features

- **Drawing Primitives:**  
  Supports a variety of commands to render:
  - **Pixel:**  
    `plot x y [colorIndex]`
  - **Line:**  
    `line x1 y1 x2 y2 [colorIndex]` and `lineto x y [colorIndex]`
  - **Circle:**  
    `circle x y radius [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)
  - **Rectangle:**  
    `rect x y width height [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)
  - **Triangle:**  
    `triangle x1 y1 x2 y2 x3 y3 [colorIndex] [mode]`  
    _Mode:_ `"F"` (fill, default) or `"S"` (stroke)

- **Colour Management:**  
  Set the drawing state with:
  - **ink:**  
    `ink color` – Sets the default ink (foreground) colour (0–7).
  - **paper:**  
    `paper color` – Sets the default paper (background) colour (0–7).
  - **bright:**  
    `bright 0|1` – Sets the brightness flag (0 for off, 1 for on).
  - **colour:**  
    `colour ink paper bright` – Sets ink, paper, and brightness all at once.

  For drawing primitives, if the color parameter is omitted (or specified as an underscore `_` to skip it when providing a later parameter), the server uses the effective default ink colour (or paper colour for background operations).

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
   git clone https://github.com/ha1tch/zxvdu.git
   cd zxvdu
   ```

2. **Install dependencies:**

   Ensure that `raylib-go` is installed:

   ```bash
   go get github.com/gen2brain/raylib-go/raylib
   ```

3. **Build the project:**

   ```bash
   go build -o zxvdu
   ```

4. **Run the server:**

   ```bash
   ./zxvdu
   ```

   You can also supply command‑line flags to set initial values:

   ```bash
   ./zxvdu -ink=3 -paper=0 -bright=1 -graphics=2 -zoom=2
   ```

   This example sets the default ink colour to 3, paper to 0, brightness on, internal resolution to 512×384 (2× multiplier), and a zoom factor of 2 (resulting in a window size of 1024×768).

---

## Usage

### Command Protocol

Commands are sent as single lines of text (terminated by a newline) with space‑separated parameters. Here are some examples:

#### Drawing Commands

- **Plot a Pixel:**  
  Use an explicit color, or omit it to use the default ink.
  - Explicit:  
    ```bash
    plot 50 50 2
    ```  
    (Draws a pixel at (50,50) using color index 2.)  
  - Omitted (defaults to effective ink):  
    ```bash
    plot 50 50
    ```

- **Draw a Line:**  
  You can either specify the color or leave it out.
  - Explicit:  
    ```bash
    line 10 10 100 100 3
    ```  
    (Draws a line from (10,10) to (100,100) using color index 3.)  
  - Omitted:  
    ```bash
    line 10 10 100 100
    ```  
    (Uses the default ink.)

- **Draw a Circle:**  
  Since the circle command accepts an optional color and an optional mode, you can provide:
  - With both color and mode specified:  
    ```bash
    circle 200 200 50 4 S
    ```  
    (Draws a stroked circle centered at (200,200) with radius 50 using color index 4.)  
  - With only color specified (mode defaults to fill):  
    ```bash
    circle 200 200 50 4
    ```  
    (Draws a filled circle with color index 4.)  
  - With both parameters omitted (using default ink and fill mode):  
    ```bash
    circle 200 200 50
    ```

- **Draw a Rectangle:**  
  Similar to circle, you can control color and mode:
  - Filled rectangle with explicit color:  
    ```bash
    rect 300 300 150 100 5
    ```  
    (Draws a filled rectangle at (300,300) of size 150×100 using color index 5.)  
  - Stroked rectangle with default ink:  
    ```bash
    rect 300 300 150 100 _ S
    ```  
    (Skips an explicit color—using the default ink—and sets mode to stroke.)  
  - Fully omitted color parameter:  
    ```bash
    rect 300 300 150 100
    ```  
    (Uses the default ink and fill mode.)

- **Draw a Triangle:**  
  Again, color and mode are optional:
  - Filled triangle with explicit color:  
    ```bash
    triangle 100 100 150 50 200 100 2
    ```  
    (Draws a filled triangle using color index 2.)  
  - Stroked triangle using default ink:  
    ```bash
    triangle 100 100 150 50 200 100 _ S
    ```  
    (Skips an explicit color so that the default ink is used, and sets mode to stroke.)  
  - With both parameters omitted:  
    ```bash
    triangle 100 100 150 50 200 100
    ```  
    (Uses the default ink and fill mode.)

#### Colour and Drawing State

- **Set Default Ink:**  
  ```bash
  ink 3
  ```  
  (Sets the default ink color to index 3.)

- **Set Default Paper:**  
  ```bash
  paper 7
  ```  
  (Sets the default paper color to index 7.)

- **Set Brightness:**  
  ```bash
  bright 1
  ```  
  (Enables brightness.)

- **Set All Colour Values at Once:**  
  ```bash
  colour 0 7 1
  ```  
  (Sets ink to 0, paper to 7, and brightness on.)

#### Buffer Commands

- **Clear Current Buffer:**  
  ```bash
  cls
  ```  
  (Clears the active buffer; the background will be painted using the effective paper color.)

- **Flip Buffers:**  
  - Toggle between buffers 0 and 1:  
    ```bash
    flip
    ```  
  - Or select a specific buffer (e.g., buffer 3):  
    ```bash
    flip 3
    ```

#### Resolution and Scaling

- **Set Graphics Resolution:**  
  This command resets all buffers and sets the internal resolution to the base (256×192) multiplied by the given factor.
  - For default resolution:  
    ```bash
    graphics 1
    ```  
    (Internal resolution becomes 256×192.)  
  - For higher resolution:  
    ```bash
    graphics 2
    ```  
    (Internal resolution becomes 512×384.)

- **Set Zoom Factor:**  
  Adjusts the display scaling (without affecting the internal buffer).
  - For no zoom (1:1 display):  
    ```bash
    zoom 1
    ```  
  - For zoom factor 2:  
    ```bash
    zoom 2
    ```  
    (If the internal resolution is, for example, 512×384, the window will scale to 1024×768 on screen.)

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
  Contains the full implementation of zxvdu including:
  - Command parsing and handling.
  - Drawing primitives rendering.
  - Buffer management.
  - Graphics resolution and zoom management.
  - Network servers for drawing commands and GUI events.
  - Command‑line flag parsing for initial settings.

---

## Getting Started on the Command Line

When starting zxvdu, you can configure the following parameters with command‑line flags:

- **-ink**: Default ink (foreground) colour (0–7).  
- **-paper**: Default paper (background) colour (0–7).  
- **-bright**: Default brightness flag (0 or 1).  
- **-graphics**: Internal resolution multiplier (>=1).  
- **-zoom**: Display zoom factor (>=1).

Example:

```bash
./zxvdu -ink=3 -paper=0 -bright=1 -graphics=2 -zoom=2
```

This sets the default ink colour to 3, paper to 0, brightness on, internal resolution to 512×384 (2× multiplier), and a zoom factor of 2 (resulting in a window size of 1024×768).

---

## Contact

**Name:** haitch  
**Email:** [haitch@duck.com](mailto:haitch@duck.com)  
**Social Media:** [https://oldbytes.space/@haitchfive](https://oldbytes.space/@haitchfive)

