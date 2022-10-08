import cv2
from rembg.bg import remove
import os


DIR = os.getcwd()
UPLOAD_FOLDER = DIR + '/uploads'
DONE_FOLDER = DIR + '/cut'

def cut(filename):
    input = cv2.imread(f"{UPLOAD_FOLDER}/{filename}")
    output = remove(input)
    cv2.imwrite(f'{DONE_FOLDER}/{filename}', output)

#
# def cut_img(filename):
#     image = cv2.imread(f"uploads/{filename}")
#     gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
#
#     gradX = cv2.Sobel(gray, cv2.CV_32F, dx=1, dy=0, ksize=-1)
#     gradY = cv2.Sobel(gray, cv2.CV_32F, dx=0, dy=1, ksize=-1)
#
#     gradient = cv2.subtract(gradX, gradY)
#     gradient = cv2.convertScaleAbs(gradient)
#
#     blurred = cv2.blur(gradient, (9, 9))
#     _, thresh = cv2.threshold(blurred, 90, 255, cv2.THRESH_BINARY)
#
#     kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (25, 25))
#     closed = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel)
#
#     closed = cv2.erode(closed, None, iterations=4)
#     closed = cv2.dilate(closed, None, iterations=4)
#
#     cnts, _b = cv2.findContours(closed.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
#
#     cv2.drawContours(image, cnts, -1, (255, 0, 0), 3, cv2.LINE_AA, _b, 1)
#     cv2.imwrite(f'cut/{filename}', image)
