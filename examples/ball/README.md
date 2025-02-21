# Animation Example

Here's a complete example showing how to create smooth animations using double buffering. This example creates a bouncing ball with realistic physics-inspired motion using easing functions.

## Python Example - Bouncing Ball

Save this as `bouncing_ball.py`:

```python
#!/usr/bin/env python3
"""
ZXVDU Bouncing Ball Example
--------------------------
Demonstrates smooth animation using double buffering and easing functions.
"""

[Previous Python code goes here]
```

### How It Works

1. **Double Buffering**
   - Uses two flip buffers (0 and 1)
   - Draws next frame to invisible buffer
   - Flips buffers to display new frame
   - Prevents visual tearing

2. **Smooth Animation**
   - Uses easing functions for natural motion
   - Bounce effect with ease-out physics
   - Horizontal motion with ease-in-out
   - Time-based animation for consistent speed

3. **VDU Features Used**
   - Buffer management (`paint`, `flip`)
   - Shape drawing (`circle`)
   - Color control (`ink`, `paper`, `bright`)
   - Screen clearing (`cls`)

### Running the Example

1. Start ZXVDU server:
   ```bash
   ./zxvdu -graphics=2 -zoom=2
   ```

2. Run the Python script:
   ```bash
   python3 bouncing_ball.py
   ```

### Expected Output

You should see:
- A red ball moving smoothly across the screen
- Natural bouncing motion with acceleration
- Smooth horizontal movement
- No visual tearing due to double buffering

### Creating Your Own Animations

Key principles to follow:
1. Always use double buffering for smooth animation
2. Use time-based movement for consistent speed
3. Clear buffers before drawing new frames
4. Aim for 60 FPS (1/60 second delay)
5. Clean up buffers when done

This example demonstrates proper usage of ZXVDU's buffer system and shows how to create smooth, professional-looking animations. The code is well-documented and can serve as a template for your own animation projects.