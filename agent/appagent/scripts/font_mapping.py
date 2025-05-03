from fontTools.ttLib import TTFont
from PIL import Image, ImageDraw, ImageFont
import json



font_file = "/home/zrik/Downloads/d24b9y.1735246212693.ttf"
font = TTFont(font_file)

cmap_table = font['cmap']

unicode_cmap = None
for table in cmap_table.tables:
    if table.isUnicode():
        unicode_cmap = table
        break

if unicode_cmap is None:
    raise ValueError("Không tìm thấy bảng cmap Unicode trong font!")

mapping = {}

for codepoint, name in unicode_cmap.cmap.items():
    if isinstance(name, str): 
        mapping[f"\\u{codepoint:04X}"] = name

def render_char(font, char):
    img = Image.new('RGB', (50, 50), (255, 255, 255))  
    draw = ImageDraw.Draw(img)
    draw.text((0, 0), char, font=font, fill=(0, 0, 0))
    return img

font = ImageFont.truetype(font_file, 40)  

for i in range(0xE000, 0xF8FF):  
    char = chr(i)
    try:
        img = render_char(font, char)
        mapping[f"\\u{i:04X}"] = char
    except:

output_file = 'letter_mapping.json'
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(mapping, f, ensure_ascii=False, indent=4)

