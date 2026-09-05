from PIL import Image

# Open image
img = Image.open('apps/web/public/zleek-logo.png').convert("RGBA")

# Assuming the icon is in the center, and the background is dark.
# Let's crop a square from the center top portion.
width, height = img.size

# Let's crop it to a 450x450 square from the center, shifted slightly up to avoid the text
# Or we can just crop out the text by taking the top 500 pixels and then center cropping
box = (width//2 - 250, height//2 - 270, width//2 + 250, height//2 + 230)
cropped = img.crop(box)

# Now let's try to make the background transparent
# We'll get the color at (0,0) of the cropped image
bg_color = cropped.getpixel((0,0))
threshold = 30

data = cropped.getdata()
new_data = []
for item in data:
    if abs(item[0]-bg_color[0]) < threshold and abs(item[1]-bg_color[1]) < threshold and abs(item[2]-bg_color[2]) < threshold:
        new_data.append((255, 255, 255, 0))
    else:
        new_data.append(item)
        
cropped.putdata(new_data)
cropped.save('apps/web/public/logo-icon.png')
cropped.save('apps/web/web/public/logo-icon.png')
print("Cropped and processed!")
