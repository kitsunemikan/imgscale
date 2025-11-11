## `imgscale`

Tiny CLI utility to scale images.

### Features

- Different scaling algorithms (nearest neighbor, linear, cubic, average color, lanczos)
- Uniform scaling by a factor
- Uniform scaling to specified target dimension
- Specify scaling factors for width and height
- Specify concrete target width and height
- Overwrite protection
- Output format autodetection
- Specify quality factor for output JPEG images

### Usage

```bash
# reencode
imgscale -i input.png -o output.jpeg -q 75

# downscale 2x + overwrite
imgscale -i input.png -o output.jpeg -s 0.5 -f

# squish horizontally 4x
imgscale -i input.png -o output2.jpeg -sx 0.25

# resize to 512x512 square
imgscale -i input.png -o output3.jpeg -ow 512 -oh 512

# resize to HD + nearest neighbor resampling
imgscale -i input.png -o output4.jpen -maxside 1280 -r nearest

# show help + available resampling algorithms
imgscale -help
```

### Install
```bash
go install github.com/kitsunemikan/imgscale@latest
```
