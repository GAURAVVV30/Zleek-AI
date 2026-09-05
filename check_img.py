from PIL import Image
import sys

img = Image.open('apps/web/public/zleek-logo.png')
print(f"Size: {img.size}")
print(f"Mode: {img.mode}")
