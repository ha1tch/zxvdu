#!/usr/bin/env python3
"""
ZXVDU Bouncing Ball Example
--------------------------
Demonstrates smooth animation using double buffering and easing functions.
The ball bounces with realistic acceleration and deceleration.

Requirements:
- Python 3.6+
- math module (standard library)
- socket module (standard library)
- time module (standard library)
"""

import math
import socket
import time
from typing import Tuple

class VDU:
    """Simple ZXVDU client implementation."""
    
    def __init__(self, host: str = "localhost", port: int = 55550):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((host, port))
        
    def send(self, cmd: str) -> str:
        """Send a command and return the response."""
        self.sock.send((cmd + "\n").encode())
        return self.sock.recv(1024).decode().strip()
        
    def close(self):
        """Close the connection."""
        self.sock.close()

def ease_out_bounce(x: float) -> float:
    """
    Bounce easing out function.
    Creates a bouncing effect that starts fast and then decelerates.
    """
    n1 = 7.5625
    d1 = 2.75

    if x < 1 / d1:
        return n1 * x * x
    elif x < 2 / d1:
        x -= 1.5 / d1
        return n1 * x * x + 0.75
    elif x < 2.5 / d1:
        x -= 2.25 / d1
        return n1 * x * x + 0.9375
    else:
        x -= 2.625 / d1
        return n1 * x * x + 0.984375

def ease_in_out_quad(x: float) -> float:
    """
    Quadratic easing in/out function.
    Accelerates until halfway, then decelerates.
    """
    return 2 * x * x if x < 0.5 else 1 - pow(-2 * x + 2, 2) / 2

def interpolate(start: float, end: float, progress: float) -> float:
    """Linear interpolation between start and end values."""
    return start + (end - start) * progress

class BouncingBall:
    def __init__(self, vdu: VDU, width: int = 256, height: int = 192):
        self.vdu = vdu
        self.width = width
        self.height = height
        
        # Ball properties
        self.radius = 10
        self.color = 2  # Red
        
        # Animation properties
        self.bounce_duration = 1.0  # seconds per bounce
        self.horizontal_duration = 2.0  # seconds to cross screen
        
        # Initialize position tracking
        self.start_time = time.time()
        
    def setup(self):
        """Set up initial VDU state."""
        # Set up double buffering - we'll use buffers 0 and 1
        self.vdu.send("paint flip")  # Select flip buffer mode
        self.vdu.send("paper 0")     # Black background
        self.vdu.send("ink 2")       # Red foreground
        self.vdu.send("bright 1")    # Bright colors
        
        # Clear both buffers
        self.vdu.send("paint 0")
        self.vdu.send("cls")
        self.vdu.send("paint 1")
        self.vdu.send("cls")
    
    def calculate_position(self, current_time: float) -> Tuple[int, int]:
        """Calculate the current ball position based on time."""
        elapsed = current_time - self.start_time
        
        # Horizontal movement (continuous back and forth)
        horizontal_progress = (elapsed % self.horizontal_duration) / self.horizontal_duration
        if int(elapsed / self.horizontal_duration) % 2:  # Reverse direction every cycle
            horizontal_progress = 1 - horizontal_progress
        horizontal_progress = ease_in_out_quad(horizontal_progress)
        x = interpolate(self.radius, self.width - self.radius, horizontal_progress)
        
        # Vertical movement (bouncing)
        bounce_progress = (elapsed % self.bounce_duration) / self.bounce_duration
        bounce_progress = ease_out_bounce(bounce_progress)
        y = interpolate(self.height - self.radius, self.radius, bounce_progress)
        
        return int(x), int(y)
    
    def draw_frame(self, buffer: int, x: int, y: int):
        """Draw a single frame to the specified buffer."""
        # Select buffer and clear it
        self.vdu.send(f"paint {buffer}")
        self.vdu.send("cls")
        
        # Draw the ball
        self.vdu.send(f"circle {x} {y} {self.radius} {self.color} F")
    
    def animate(self, duration: float = 10.0):
        """Run the animation for the specified duration."""
        try:
            self.setup()
            
            # Animation loop
            active_buffer = 0
            end_time = time.time() + duration
            
            while time.time() < end_time:
                # Calculate new position
                x, y = self.calculate_position(time.time())
                
                # Draw to back buffer
                back_buffer = 1 - active_buffer
                self.draw_frame(back_buffer, x, y)
                
                # Flip buffers
                self.vdu.send(f"flip {back_buffer}")
                active_buffer = back_buffer
                
                # Control frame rate
                time.sleep(1/60)  # Aim for 60 FPS
                
        except KeyboardInterrupt:
            pass
        finally:
            # Clean up - clear both buffers
            self.vdu.send("paint 0")
            self.vdu.send("cls")
            self.vdu.send("paint 1")
            self.vdu.send("cls")

def main():
    # Create VDU connection
    vdu = VDU()
    try:
        # Create and run animation
        ball = BouncingBall(vdu)
        ball.animate()
    finally:
        vdu.close()

if __name__ == "__main__":
    main()