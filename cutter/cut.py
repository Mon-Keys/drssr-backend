import cv2
from rembg.bg import remove
import os

DIR = os.getcwd()
UPLOAD_FOLDER = DIR + '/clothes'
DONE_FOLDER = DIR + '/masks'

def cut(filename):
    input = cv2.imread(f"{UPLOAD_FOLDER}/{filename}")
    output = remove(input)
    filename1 = filename.split('.')[0] + '.png'
    cv2.imwrite(f'{DONE_FOLDER}/{filename1}', output)
    return filename1