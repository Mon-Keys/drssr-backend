import cv2
from rembg.bg import remove

def cut(img_path, mask_path):
    input = cv2.imread(img_path)
    output = remove(input)
    cv2.imwrite(mask_path, output)
    return
