"""
ZXVDU Python Client Library
--------------------------
A Python interface for the ZXVDU display server.

This module provides a high-level interface for working with ZXVDU,
including buffer management, texture handling, and drawing operations.
"""

import socket
import time
from dataclasses import dataclass
from enum import Enum, auto
from typing import List, Tuple, Optional, Union
import math

class VDUError(Exception):
    """Base exception for VDU errors."""
    pass

class TextureError(VDUError):
    """Exception for texture-related errors."""
    pass

class CommandError(VDUError):
    """Exception for command execution errors."""
    pass

class DrawMode(Enum):
    """Drawing modes for shapes."""
    FILL = "F"
    STROKE = "S"
    TEXTURE = "T"

class BufferMode(Enum):
    """Buffer selection modes."""
    FLIP = "flip"
    LAYER = "layer"

@dataclass
class Color:
    """Predefined ZX Spectrum colors."""
    BLACK = 0
    BLUE = 1
    RED = 2
    MAGENTA = 3
    GREEN = 4
    CYAN = 5
    YELLOW = 6
    WHITE = 7

    @staticmethod
    def bright(color: int) -> int:
        """Convert a color to its bright variant."""
        if color == 0:
            return 0
        return color + 7

class Texture:
    """Represents a captured texture in ZXVDU."""
    def __init__(self, vdu: 'VDU', slot: int, width: int, height: int):
        self.vdu = vdu
        self.slot = slot
        self.width = width
        self.height = height
        self._valid = True

    def draw(self, x: int, y: int):
        """Draw the texture at the specified position."""
        if not self._valid:
            raise TextureError("Texture has been deleted")
        self.vdu._send(f"tex paint {x} {y} {self.slot}")

    def delete(self):
        """Delete the texture, freeing its slot."""
        if self._valid:
            self.vdu._send(f"tex del {self.slot}")
            self._valid = False

    def __del__(self):
        """Ensure texture is deleted when object is garbage collected."""
        try:
            self.delete()
        except:
            pass

class VDU:
    """Main interface to ZXVDU server."""
    
    def __init__(self, host: str = "localhost", port: int = 55550):
        """Initialize connection to ZXVDU server."""
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            self.sock.connect((host, port))
        except ConnectionRefusedError:
            raise VDUError(f"Could not connect to ZXVDU server at {host}:{port}")
        
        # Initialize default state
        self.current_color = Color.WHITE
        self.current_mode = BufferMode.FLIP
        self.bright = False
        
    def _send(self, cmd: str) -> str:
        """Send a command and return the response."""
        try:
            self.sock.send((cmd + "\n").encode())
            response = self.sock.recv(1024).decode().strip()
            
            if response.startswith("ERROR"):
                raise CommandError(response)
            
            return response
        except socket.error as e:
            raise VDUError(f"Communication error: {e}")

    def close(self):
        """Close the connection to ZXVDU."""
        try:
            self.sock.close()
        except:
            pass

    def select_buffer(self, mode: BufferMode, number: Optional[int] = None):
        """Select buffer mode and optionally a specific buffer."""
        self.current_mode = mode
        self._send(f"paint {mode.value}")
        if number is not None:
            if mode == BufferMode.FLIP:
                self._send(f"flip {number}")
            else:
                self._send(f"layer {number}")

    def set_color(self, color: int, bright: bool = False):
        """Set the current drawing color."""
        self.current_color = color
        self.bright = bright
        self._send(f"ink {color}")
        self._send(f"bright {1 if bright else 0}")

    def clear(self):
        """Clear the current buffer."""
        self._send("cls")

    def plot(self, x: int, y: int, color: Optional[int] = None):
        """Plot a single pixel."""
        color_str = f" {color}" if color is not None else ""
        self._send(f"plot {x} {y}{color_str}")

    def line(self, x1: int, y1: int, x2: int, y2: int, color: Optional[int] = None):
        """Draw a line between two points."""
        color_str = f" {color}" if color is not None else ""
        self._send(f"line {x1} {y1} {x2} {y2}{color_str}")

    def rect(self, x: int, y: int, width: int, height: int, 
             color: Optional[int] = None, mode: DrawMode = DrawMode.FILL) -> Optional[Texture]:
        """
        Draw or capture a rectangle.
        Returns a Texture object if mode is TEXTURE, None otherwise.
        """
        color_str = f" {color}" if color is not None else ""
        response = self._send(f"rect {x} {y} {width} {height}{color_str} {mode.value}")
        
        if mode == DrawMode.TEXTURE:
            try:
                slot = int(response)
                return Texture(self, slot, width, height)
            except ValueError:
                raise TextureError("Failed to capture texture")
        return None

    def circle(self, x: int, y: int, radius: int, 
               color: Optional[int] = None, mode: DrawMode = DrawMode.FILL):
        """Draw a circle."""
        color_str = f" {color}" if color is not None else ""
        self._send(f"circle {x} {y} {radius}{color_str} {mode.value}")

    def triangle(self, x1: int, y1: int, x2: int, y2: int, x3: int, y3: int,
                color: Optional[int] = None, mode: DrawMode = DrawMode.FILL):
        """Draw a triangle."""
        color_str = f" {color}" if color is not None else ""
        self._send(f"triangle {x1} {y1} {x2} {y2} {x3} {y3}{color_str} {mode.value}")

    def create_texture_from_data(self, data: str, width: int, height: int) -> Texture:
        """Create a texture from hex pixel data."""
        response = self._send(f"tex add {data} {width} {height}")
        try:
            slot = int(response)
            return Texture(self, slot, width, height)
        except ValueError:
            raise TextureError("Failed to create texture")

    def __enter__(self):
        """Support for context manager protocol."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Ensure connection is closed when used as context manager."""
        self.close()
