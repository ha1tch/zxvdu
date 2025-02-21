# Changelog
All notable changes to ZXVDU will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-02-21
### Added
- New texture capture command using "rect x y width height T"
- Case-insensitive handling for all shape modes (F/S/T)
- Mouse coordinates now properly scaled with zoom factor
- Graphics resolution multiplier via -graphics flag
- Display zoom factor via -zoom flag
- Better error handling with standardized error codes
- String payload support in command structure
- Unified texture operation handling
- Comprehensive command documentation (COMMANDS.md)

### Changed
- Switched to new BufferSystem for all buffer operations
- Simplified buffer management with consistent flip/layer approach
- Consolidated texture handling between pixel data and buffer capture
- Standardized error message format (ERROR XXXX : message)
- Improved memory management for textures
- Updated mouse event coordinates to account for zoom factor
- Unified texture command responses
- Better validation for texture operations

### Fixed
- Texture slot management memory leaks
- Proper cleanup of unused textures
- Buffer synchronization issues
- Mouse coordinate accuracy at different zoom levels
- Missing string payload in DrawCommand structure
- Inconsistent texture error handling
- Command parsing edge cases
- Resource cleanup in error scenarios

### Breaking Changes
- paint_target and paint_copy commands removed
- Buffer selection now uses paint flip/layer syntax
- Texture capture uses new rect T syntax
- Error response format changed to ERROR XXXX : message
- Texture command responses standardized

### Migration Guide
- Replace paint_target onscreen with paint flip
- Replace paint_target offscreen with paint layer
- Replace paint_copy with appropriate buffer selection
- Update error handling to expect new format
- Use rect T for texture capture operations

## [0.1.0] - 2024-12-15
- Initial pre-release
- Basic drawing commands
- Network command interface
- Simple buffer system
- Initial texture support

## Notes
- This is pre-release software; APIs may change without notice
- The README.md documentation needs to be updated to reflect these changes
- Legacy scripts using paint_target or paint_copy will need updates
- Custom clients may need updates to handle new error format