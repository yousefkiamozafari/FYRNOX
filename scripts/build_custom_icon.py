import os
import shutil
import collections
from PIL import Image, ImageFilter, ImageEnhance

def extract_transparent_fox(src_path):
    print(f"Loading source image: {src_path}")
    img = Image.open(src_path).convert("RGBA")
    w, h = img.size

    # BFS flood fill from all perimeter edges to find pure black background
    visited = bytearray(w * h)
    queue = collections.deque()

    for x in range(w):
        queue.append((x, 0))
        queue.append((x, h - 1))
        visited[x] = 1
        visited[(h - 1) * w + x] = 1

    for y in range(h):
        queue.append((0, y))
        queue.append((w - 1, y))
        visited[y * w] = 1
        visited[y * w + (w - 1)] = 1

    threshold = 32  # Anything <= threshold in R+G+B is background
    while queue:
        x, y = queue.popleft()
        for nx, ny in [(x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)]:
            if 0 <= nx < w and 0 <= ny < h:
                nidx = ny * w + nx
                if not visited[nidx]:
                    r, g, b, _ = img.getpixel((nx, ny))
                    if (r + g + b) <= threshold:
                        visited[nidx] = 1
                        queue.append((nx, ny))

    # Build alpha mask: 0 for background, 255 for fox
    mask = Image.new("L", (w, h), 255)
    for y in range(h):
        for x in range(w):
            if visited[y * w + x]:
                mask.putpixel((x, y), 0)

    # Smooth mask border slightly to eliminate any pixelated stair-stepping
    mask_smooth = mask.filter(ImageFilter.GaussianBlur(radius=0.7))

    # Apply alpha mask
    fox_rgba = img.copy()
    fox_rgba.putalpha(mask_smooth)

    # Clean up any near-zero alpha noise at outer corners
    for y in range(h):
        for x in range(w):
            if visited[y * w + x]:
                r, g, b, a = fox_rgba.getpixel((x, y))
                if a < 15:
                    fox_rgba.putpixel((x, y), (0, 0, 0, 0))

    # Crop tightly to fox bounding box
    bbox = fox_rgba.getbbox()
    print(f"Tight bounding box: {bbox} (width: {bbox[2] - bbox[0]}, height: {bbox[3] - bbox[1]})")
    cropped = fox_rgba.crop(bbox)

    # Center in a clean square canvas with ~4% breathing margin
    cw, ch = cropped.size
    dim = max(cw, ch)
    canvas_dim = int(dim * 1.08)  # ~4% padding on each side
    canvas = Image.new("RGBA", (canvas_dim, canvas_dim), (0, 0, 0, 0))
    px = (canvas_dim - cw) // 2
    py = (canvas_dim - ch) // 2
    canvas.paste(cropped, (px, py), mask=cropped)

    # Enhance color vibrancy slightly for crisp, radiant icon visibility
    enhancer = ImageEnhance.Color(canvas)
    canvas = enhancer.enhance(1.15)
    contrast = ImageEnhance.Contrast(canvas)
    canvas = contrast.enhance(1.06)

    return canvas

def main():
    repo_root = r"C:\Users\LouiNoor\.gemini\antigravity\scratch\FoxyVPN-Windows"
    src_path = r"C:\Users\LouiNoor\.gemini\antigravity\brain\e4a584ef-f1a1-440a-ab06-38e888100028\.user_uploaded\media_1792873823001_be8d5a62.jpg"

    os.chdir(repo_root)
    canvas = extract_transparent_fox(src_path)

    # 1. Save 512x512 icon.png (Transparent background)
    png512 = canvas.resize((512, 512), Image.Resampling.LANCZOS)
    png512.save("icon.png", format="PNG", optimize=True)
    os.makedirs(r"cmd\fyrnox\web", exist_ok=True)
    png512.save(r"cmd\fyrnox\web\favicon.png", format="PNG", optimize=True)
    print("Saved transparent icon.png and cmd/fyrnox/web/favicon.png (512x512)")

    # 2. Build multi-resolution ICO (256, 128, 64, 48, 32, 24, 16) with clean alpha
    ico_sizes = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (24, 24), (16, 16)]
    layers = []
    for s in ico_sizes:
        layer = canvas.resize(s, Image.Resampling.LANCZOS)
        # Clean alpha fringe
        r, g, b, a = layer.split()
        a = a.point(lambda p: 0 if p < 10 else p)
        layer.putalpha(a)
        layers.append(layer)

    layers[0].save("icon.ico", format="ICO", append_images=layers[1:])
    print("Saved multi-resolution transparent icon.ico with sizes:", ico_sizes)

    # 3. Copy icon.ico to all target locations
    targets = [
        r"cmd\fyrnox\web\favicon.ico",
        r"cmd\installer\payload\icon.ico",
        r"build\portable\Fyrnox-Portable\icon.ico",
        r"build\portable-x86\Fyrnox-Portable\icon.ico",
    ]
    for target in targets:
        dir_name = os.path.dirname(target)
        if dir_name:
            os.makedirs(dir_name, exist_ok=True)
        shutil.copyfile("icon.ico", target)
        print(f"Copied icon.ico to {target}")

if __name__ == "__main__":
    main()
