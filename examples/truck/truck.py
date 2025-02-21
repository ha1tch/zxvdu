#!/usr/bin/env python3
"""
ZXVDU Moving Truck Example
-------------------------
Demonstrates layer usage, texture capture, and buffer swapping.
Creates a scene with a moving background and an animated truck.

Features demonstrated:
- Layer compositing
- Texture capture and reuse
- Buffer swapping for animation
- Background scrolling
- Multiple buffer management
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

class TruckAnimation:
    def __init__(self, vdu: VDU, width: int = 256, height: int = 192):
        self.vdu = vdu
        self.width = width
        self.height = height
        
        # Animation properties
        self.scroll_speed = 2  # pixels per frame
        self.wheel_frames = 4  # number of wheel animation frames
        self.current_wheel = 0
        self.wheel_delay = 0.1  # seconds between wheel frames
        self.last_wheel_update = 0
        
        # Background properties
        self.mountain_height = 40
        self.ground_height = 40
        
        # Truck position
        self.truck_x = 80
        self.truck_y = height - self.ground_height - 30
        
        # Storage for texture slots
        self.wheel_textures = []
        self.truck_body_texture = None
        
    def draw_mountain_background(self, buffer: int):
        """Draw a simple mountain landscape."""
        self.vdu.send(f"paint {buffer}")
        self.vdu.send("cls")
        
        # Sky
        self.vdu.send("ink 1")  # Blue
        self.vdu.send(f"rect 0 0 {self.width} {self.height - self.ground_height} _ F")
        
        # Mountains
        self.vdu.send("ink 4")  # Green
        points = [
            (0, self.height - self.ground_height),
            (50, self.height - self.ground_height - self.mountain_height),
            (100, self.height - self.ground_height),
            (150, self.height - self.ground_height - self.mountain_height * 0.7),
            (200, self.height - self.ground_height),
            (250, self.height - self.ground_height - self.mountain_height * 0.5),
            (self.width, self.height - self.ground_height)
        ]
        
        for i in range(len(points) - 1):
            x1, y1 = points[i]
            x2, y2 = points[i + 1]
            self.vdu.send(f"triangle {x1} {y1} {x2} {y2} {x2} {y1} _ F")
        
        # Ground
        self.vdu.send("ink 6")  # Yellow
        self.vdu.send(f"rect 0 {self.height - self.ground_height} {self.width} {self.ground_height} _ F")
    
    def create_truck_textures(self):
        """Create textures for truck body and wheels."""
        # Draw and capture truck body
        self.vdu.send("paint 1")
        self.vdu.send("cls")
        
        # Main body (red)
        self.vdu.send("ink 2")
        self.vdu.send(f"rect {self.truck_x} {self.truck_y} 60 20 _ F")
        self.vdu.send(f"rect {self.truck_x + 40} {self.truck_y - 15} 20 15 _ F")
        
        # Windows (cyan)
        self.vdu.send("ink 5")
        self.vdu.send(f"rect {self.truck_x + 42} {self.truck_y - 12} 15 8 _ F")
        
        # Capture truck body
        response = self.vdu.send(f"rect {self.truck_x} {self.truck_y - 15} 60 35 T")
        self.truck_body_texture = int(response)
        
        # Create wheel frames
        wheel_y = self.truck_y + 15
        for i in range(self.wheel_frames):
            self.vdu.send("paint 1")
            self.vdu.send("cls")
            
            # Draw wheel with spokes at different angles
            self.vdu.send("ink 7")  # White
            angle = (i / self.wheel_frames) * 2 * math.pi
            cx, cy = 10, 10  # Center of wheel
            radius = 8
            
            # Wheel rim
            self.vdu.send(f"circle {cx} {cy} {radius} _ S")
            
            # Spokes
            for spoke in range(4):
                spoke_angle = angle + (spoke * math.pi / 2)
                x = cx + math.cos(spoke_angle) * radius
                y = cy + math.sin(spoke_angle) * radius
                self.vdu.send(f"line {cx} {cy} {int(x)} {int(y)} _")
            
            # Capture wheel frame
            response = self.vdu.send(f"rect 0 0 20 20 T")
            self.wheel_textures.append(int(response))
    
    def draw_background(self, offset: int):
        """Draw scrolling background in flip buffer."""
        # Draw base background twice to allow scrolling
        for i in range(2):
            pos_x = (i * self.width) - offset
            if pos_x < self.width:
                self.draw_mountain_background(0)
    
    def draw_truck(self):
        """Draw truck with current wheel frame in layer buffer."""
        self.vdu.send("paint layer")
        self.vdu.send("cls")
        
        # Draw truck body
        self.vdu.send(f"tex paint {self.truck_x} {self.truck_y - 15} {self.truck_body_texture}")
        
        # Draw wheels
        wheel_texture = self.wheel_textures[self.current_wheel]
        self.vdu.send(f"tex paint {self.truck_x + 10} {self.truck_y + 10} {wheel_texture}")
        self.vdu.send(f"tex paint {self.truck_x + 40} {self.truck_y + 10} {wheel_texture}")
    
    def update_wheel_frame(self):
        """Update wheel animation frame if enough time has passed."""
        current_time = time.time()
        if current_time - self.last_wheel_update >= self.wheel_delay:
            self.current_wheel = (self.current_wheel + 1) % self.wheel_frames
            self.last_wheel_update = current_time
    
    def animate(self, duration: float = 10.0):
        """Run the animation for the specified duration."""
        try:
            # Initial setup
            offset = 0
            self.create_truck_textures()
            
            # Animation loop
            start_time = time.time()
            while time.time() - start_time < duration:
                # Update background position
                offset = (offset + self.scroll_speed) % self.width
                self.draw_background(offset)
                
                # Update and draw truck
                self.update_wheel_frame()
                self.draw_truck()
                
                # Control frame rate
                time.sleep(1/30)  # Aim for 30 FPS
                
        except KeyboardInterrupt:
            pass
        finally:
            # Cleanup - clear both buffers
            self.vdu.send("paint flip")
            self.vdu.send("cls")
            self.vdu.send("paint layer")
            self.vdu.send("cls")
            
            # Clean up textures
            for texture in self.wheel_textures:
                self.vdu.send(f"tex del {texture}")
            if self.truck_body_texture is not None:
                self.vdu.send(f"tex del {self.truck_body_texture}")

def main():
    # Create VDU connection
    vdu = VDU()
    try:
        # Create and run animation
        truck = TruckAnimation(vdu)
        truck.animate()
    finally:
        vdu.close()

if __name__ == "__main__":
    main()