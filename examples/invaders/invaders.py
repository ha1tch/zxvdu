#!/usr/bin/env python3
"""
ZXVDU Space Invaders Demo
------------------------
A simple Space Invaders style demo using the ZXVDU Python module.
Demonstrates sprite animation, collision detection, and game state management.
"""

import sys
import time
import math
from dataclasses import dataclass
from typing import List, Optional, Tuple
import random
import keyboard

# Add the lib directory to the path
sys.path.append("../lib")
from zxvdu import VDU, Color, BufferMode, DrawMode, Texture, VDUError

@dataclass
class GameObject:
    """Base class for game objects."""
    x: float
    y: float
    width: int
    height: int
    texture: Optional[Texture] = None

    def intersects(self, other: 'GameObject') -> bool:
        """Check if this object intersects with another."""
        return (self.x < other.x + other.width and
                self.x + self.width > other.x and
                self.y < other.y + other.height and
                self.y + self.height > other.y)

class Invader(GameObject):
    """Represents an alien invader."""
    def __init__(self, x: float, y: float, textures: List[Texture]):
        super().__init__(x, y, textures[0].width, textures[0].height)
        self.textures = textures
        self.current_frame = 0
        self.frame_time = 0
        self.frame_delay = 0.5  # seconds between frames
        self.direction = 1  # 1 = right, -1 = left
        self.speed = 30  # pixels per second
        self.step_down = False
        
    def update(self, dt: float, bounds: Tuple[int, int]):
        """Update invader position and animation."""
        self.frame_time += dt
        if self.frame_time >= self.frame_delay:
            self.frame_time = 0
            self.current_frame = (self.current_frame + 1) % len(self.textures)
            
        # Movement
        if self.step_down:
            self.y += 10
            self.direction *= -1
            self.step_down = False
        else:
            self.x += self.direction * self.speed * dt
            
        # Check bounds
        if self.x <= bounds[0] or self.x + self.width >= bounds[1]:
            self.step_down = True

    def draw(self, vdu: VDU):
        """Draw the invader using current animation frame."""
        self.textures[self.current_frame].draw(int(self.x), int(self.y))

class Player(GameObject):
    """Represents the player's ship."""
    def __init__(self, x: float, y: float, texture: Texture):
        super().__init__(x, y, texture.width, texture.height, texture)
        self.speed = 100  # pixels per second
        self.cooldown = 0
        self.fire_rate = 0.5  # seconds between shots

    def update(self, dt: float, move_dir: int, bounds: Tuple[int, int]):
        """Update player position and shooting cooldown."""
        self.x += move_dir * self.speed * dt
        self.x = max(bounds[0], min(self.x, bounds[1] - self.width))
        
        if self.cooldown > 0:
            self.cooldown = max(0, self.cooldown - dt)

    def draw(self, vdu: VDU):
        """Draw the player's ship."""
        self.texture.draw(int(self.x), int(self.y))

class Projectile(GameObject):
    """Represents a projectile (bullet)."""
    def __init__(self, x: float, y: float, speed: float):
        super().__init__(x, y, 2, 6)
        self.speed = speed
    
    def update(self, dt: float) -> bool:
        """Update projectile position. Returns True if still active."""
        self.y += self.speed * dt
        return 0 <= self.y <= 192
    
    def draw(self, vdu: VDU):
        """Draw the projectile."""
        vdu.rect(int(self.x), int(self.y), self.width, self.height, 
                 Color.YELLOW, DrawMode.FILL)

class Game:
    """Main game class."""
    def __init__(self, vdu: VDU):
        self.vdu = vdu
        self.width = 256
        self.height = 192
        self.running = True
        self.score = 0
        
        # Create textures
        self.invader_textures = self._create_invader_textures()
        self.player_texture = self._create_player_texture()
        
        # Create game objects
        self.player = Player(self.width/2, self.height-20, self.player_texture)
        self.invaders = self._create_invaders()
        self.projectiles: List[Projectile] = []
        
        # Game state
        self.move_dir = 0
        self.shooting = False
        
    def _create_invader_textures(self) -> List[Texture]:
        """Create alien invader animation frames."""
        textures = []
        
        # Frame 1
        self.vdu.select_buffer(BufferMode.LAYER)
        self.vdu.clear()
        self.vdu.set_color(Color.GREEN, True)
        # Main body
        self.vdu.rect(2, 2, 12, 8, mode=DrawMode.FILL)
        # Tentacles
        self.vdu.rect(0, 4, 2, 2, mode=DrawMode.FILL)
        self.vdu.rect(14, 4, 2, 2, mode=DrawMode.FILL)
        textures.append(self.vdu.rect(0, 0, 16, 12, mode=DrawMode.TEXTURE))
        
        # Frame 2
        self.vdu.clear()
        self.vdu.set_color(Color.GREEN, True)
        # Main body
        self.vdu.rect(2, 2, 12, 8, mode=DrawMode.FILL)
        # Tentacles (different position)
        self.vdu.rect(0, 2, 2, 2, mode=DrawMode.FILL)
        self.vdu.rect(14, 2, 2, 2, mode=DrawMode.FILL)
        textures.append(self.vdu.rect(0, 0, 16, 12, mode=DrawMode.TEXTURE))
        
        return textures
    
    def _create_player_texture(self) -> Texture:
        """Create player ship texture."""
        self.vdu.select_buffer(BufferMode.LAYER)
        self.vdu.clear()
        self.vdu.set_color(Color.CYAN, True)
        # Ship body
        self.vdu.rect(2, 4, 12, 8, mode=DrawMode.FILL)
        # Ship nose
        self.vdu.triangle(7, 0, 2, 4, 12, 4, mode=DrawMode.FILL)
        return self.vdu.rect(0, 0, 16, 12, mode=DrawMode.TEXTURE)
    
    def _create_invaders(self) -> List[Invader]:
        """Create initial set of invaders."""
        invaders = []
        for row in range(3):
            for col in range(8):
                x = 20 + col * 24
                y = 20 + row * 20
                invaders.append(Invader(x, y, self.invader_textures))
        return invaders

    def handle_input(self):
        """Process keyboard input."""
        self.move_dir = 0
        if keyboard.is_pressed('left'):
            self.move_dir = -1
        if keyboard.is_pressed('right'):
            self.move_dir = 1
            
        if keyboard.is_pressed('space') and self.player.cooldown <= 0:
            self.player.cooldown = self.player.fire_rate
            # Create new projectile
            x = self.player.x + self.player.width/2 - 1
            y = self.player.y
            self.projectiles.append(Projectile(x, y, -150))

    def update(self, dt: float):
        """Update game state."""
        # Update player
        self.player.update(dt, self.move_dir, (0, self.width))
        
        # Update invaders
        for invader in self.invaders:
            invader.update(dt, (0, self.width))
            
        # Update projectiles and check collisions
        self.projectiles = [p for p in self.projectiles if p.update(dt)]
        
        # Check collisions
        for projectile in self.projectiles[:]:
            for invader in self.invaders[:]:
                if projectile.intersects(invader):
                    self.projectiles.remove(projectile)
                    self.invaders.remove(invader)
                    self.score += 100
                    break
        
        # Check game over conditions
        if not self.invaders:
            print(f"You win! Score: {self.score}")
            self.running = False
        elif any(invader.y + invader.height >= self.player.y for invader in self.invaders):
            print(f"Game Over! Score: {self.score}")
            self.running = False

    def draw(self):
        """Draw current game state."""
        # Clear both buffers
        self.vdu.select_buffer(BufferMode.FLIP)
        self.vdu.clear()
        self.vdu.select_buffer(BufferMode.LAYER)
        self.vdu.clear()
        
        # Draw background stars
        self.vdu.select_buffer(BufferMode.FLIP)
        self.vdu.set_color(Color.WHITE, True)
        for _ in range(20):
            x = random.randint(0, self.width-1)
            y = random.randint(0, self.height-1)
            self.vdu.plot(x, y)
            
        # Draw game objects
        self.vdu.select_buffer(BufferMode.LAYER)
        self.player.draw(self.vdu)
        for invader in self.invaders:
            invader.draw(self.vdu)
        for projectile in self.projectiles:
            projectile.draw(self.vdu)

    def run(self):
        """Main game loop."""
        last_time = time.time()
        try:
            while self.running:
                current_time = time.time()
                dt = current_time - last_time
                last_time = current_time
                
                self.handle_input()
                self.update(dt)
                self.draw()
                
                # Control frame rate
                time.sleep(max(0, 1/60 - dt))
                
        finally:
            # Cleanup
            for texture in self.invader_textures:
                texture.delete()
            self.player_texture.delete()

def main():
    try:
        with VDU() as vdu:
            game = Game(vdu)
            game.run()
    except VDUError as e:
        print(f"VDU Error: {e}")
        return 1
    except KeyboardInterrupt:
        return 0

if __name__ == "__main__":
    sys.exit(main())