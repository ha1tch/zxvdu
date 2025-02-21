# ZXVDU Space Invaders Demo

A simple Space Invaders style game demonstrating the use of the ZXVDU Python module.
Shows sprite animation, texture management, buffer compositing, and game state management.

## Requirements
- Python 3.6 or later
- keyboard module (`pip install keyboard`)
- ZXVDU Python module (../lib/zxvdu.py)

## Features Demonstrated

### Buffer Management
- Flip buffer for background stars
- Layer buffer for game sprites
- Double buffering for smooth animation

### Texture Operations
- Sprite creation and capture
- Animation frame management
- Proper texture cleanup

### Input Handling
- Keyboard input processing
- Movement control
- Shooting mechanics

### Game Mechanics
- Collision detection
- Score tracking
- Game state management

## Running the Demo

1. Start ZXVDU server:
```bash
./zxvdu -graphics=2 -zoom=2
```

2. Run the game:
```bash
python3 invaders.py
```

## Controls
- Left Arrow: Move left
- Right Arrow: Move right
- Space: Fire

## Implementation Details

### Game Objects
- `GameObject`: Base class with position and collision detection
- `Invader`: Alien invader with animation
- `Player`: Player's ship with movement and shooting
- `Projectile`: Bullets fired by the player

### Graphics
- Background stars created randomly each frame
- Invaders animated with two alternating frames
- Smooth movement with delta time
- Layer compositing for clean rendering

### Resource Management
- Proper texture cleanup on exit
- Buffer state management
- Error handling for VDU operations

## Code Structure

The code is organized into several classes:

1. Game Objects:
   - Base `GameObject` class for common functionality
   - Specialized classes for different game entities
   - Collision detection methods

2. Main Game Class:
   - Texture creation and management
   - Game loop control
   - State updates and rendering
   - Input processing

3. Resource Management:
   - Context manager usage for VDU connection
   - Proper cleanup of textures
   - Error handling for VDU operations

## Learning Points

1. Sprite Management
   - Creating sprites with basic shapes
   - Capturing sprites as textures
   - Managing animation frames

2. Buffer Usage
   - Using flip buffer for background
   - Using layer buffer for sprites
   - Proper buffer clearing and selection

3. Game Development
   - Game loop implementation
   - Delta time for smooth movement
   - Collision detection
   - State management

4. Error Handling
   - VDU connection management
   - Resource cleanup
   - Error reporting

This example serves as a template for creating more complex games using ZXVDU's
features.