import os
import shutil
from PIL import Image, ImageDraw

def create_rounded_mask(size, radius):
    mask = Image.new("L", size, 0)
    draw = ImageDraw.Draw(mask)
    draw.rounded_rectangle([(0, 0), size], radius=radius, fill=255)
    return mask

def main():
    src_path = r"C:\Users\LouiNoor\.gemini\antigravity\brain\e4a584ef-f1a1-440a-ab06-38e888100028\.user_uploaded\media_1792873823001_be8d5a62.jpg"
    print(f"Loading source image: {src_path}")
    img = Image.open(src_path).convert("RGBA")
    w, h = img.size

    # Crop to square if needed
    min_dim = min(w, h)
    left = (w - min_dim) // 2
    top = (h - min_dim) // 2
    img = img.crop((left, top, left + min_dim, top + min_dim))
    
    # Resize to standard high-res canvas 1024x1024
    img = img.resize((1024, 1024), Image.Resampling.LANCZOS)
    
    # Create smooth rounded badge mask (radius = 180 for 1024, ~17.5% squircle)
    radius = int(1024 * 0.18)
    mask = create_rounded_mask((1024, 1024), radius)
    badge_img = Image.new("RGBA", (1024, 1024), (0, 0, 0, 0))
    badge_img.paste(img, (0, 0), mask=mask)

    # 1. Save 512x512 icon.png and favicon.png
    png512 = badge_img.resize((512, 512), Image.Resampling.LANCZOS)
    png512.save("icon.png", format="PNG", optimize=True)
    os.makedirs(r"cmd\fyrnox\web", exist_ok=True)
    png512.save(r"cmd\fyrnox\web\favicon.png", format="PNG", optimize=True)
    print("Saved icon.png and cmd/fyrnox/web/favicon.png (512x512)")

    # 2. Build multi-resolution ICO (256, 128, 64, 48, 32, 24, 16)
    ico_sizes = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (24, 24), (16, 16)]
    badge_img.save("icon.ico", format="ICO", sizes=ico_sizes)
    print("Saved icon.ico with sizes:", ico_sizes)

    # 3. Copy to installer payload
    os.makedirs(r"cmd\installer\payload", exist_ok=True)
    shutil.copyfile("icon.ico", r"cmd\installer\payload\icon.ico")
    print("Copied icon.ico to cmd/installer/payload/icon.ico")

if __name__ == "__main__":
    main()
