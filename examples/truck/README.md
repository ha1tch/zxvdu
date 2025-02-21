# ZXVDU Truck Animation Example

This example demonstrates advanced ZXVDU features by creating an animated scene with a truck moving through a scrolling landscape. It showcases layer compositing, texture capture and reuse, and buffer management.

## Features Demonstrated

### Layer Management
- Background drawn in flip buffer
- Truck drawn in layer buffer
- Buffer clearing and swapping

### Texture Operations
- Texture capture from screen regions
- Texture reuse for animation
- Proper texture cleanup

### Animation Techniques
- Scrolling background
- Rotating wheel spokes
- Frame timing control

### Buffer Usage
- Flip buffer for background
- Layer buffer for truck
- Double drawing for smooth scrolling

## Running the Example

1. Start ZXVDU server:
```bash
./zxvdu -graphics=2 -zoom=2
```

2. Run the example:
```bash
python3 truck.py
```

## How It Works

### Background Scrolling
The example creates a mountain landscape that repeats horizontally. The background
is drawn twice and scrolled using an offset, creating a smooth infinite scrolling
effect.

### Truck Animation
1. The truck body is drawn once and captured as a texture
2. Multiple wheel frames are created and captured as textures
3. The wheels are animated by cycling through the wheel textures
4. The truck is drawn in the layer buffer, composited over the background

### Resource Management
The example demonstrates proper resource management:
- Textures are properly created and deleted
- Buffers are cleared on exit
- Network connection is properly closed

## Code Structure

- `VDU` class: Handles network communication with ZXVDU
- `TruckAnimation` class:
  - `draw_mountain_background`: Creates the scrolling landscape
  - `create_truck_textures`: Generates and captures truck components
  - `draw_background`: Handles background scrolling
  - `draw_truck`: Composites the truck using textures
  - `update_wheel_frame`: Manages wheel animation timing
  - `animate`: Main animation loop

## Learning Points

1. Layer Usage
   - Flip buffer for static/scrolling content
   - Layer buffer for overlaid animations
   - Buffer clearing for clean animation

2. Texture Management
   - Creating textures from screen regions
   - Storing and reusing textures
   - Proper cleanup of texture resources

3. Animation Techniques
   - Frame timing control
   - Smooth scrolling implementation
   - Component-based animation

4. Error Handling
   - Proper resource cleanup
   - Connection management
   - Exception handling

This example serves as a template for creating more complex animations using
ZXVDU's compositing and texture features.