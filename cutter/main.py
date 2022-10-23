import os
from flask import Flask, abort, flash, request, redirect, url_for, send_from_directory, jsonify
from werkzeug.utils import secure_filename
from cut import cut
import base64

DIR = os.getcwd()
UPLOAD_FOLDER = DIR + '/clothes'
DONE_FOLDER = DIR + '/masks'
ALLOWED_EXTENSIONS = {'txt', 'pdf', 'png', 'jpg', 'jpeg', 'gif', 'webp'}

app = Flask(__name__)
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
app.config['DONE_FOLDER'] = DONE_FOLDER
app.secret_key = "secret"

def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


@app.route('/upload', methods=['GET', 'POST'])
def upload_file():
    if request.method == 'POST':
        if 'file' not in request.files:
            return 'bad request', 400

        file = request.files['file']

        if file.filename == '':
            return 'bad request', 400

        if file and allowed_file(file.filename):
            filename = secure_filename(file.filename)
            file.save(os.path.join(app.config['UPLOAD_FOLDER'], filename))

            fname = cut(filename)  # Super Магия

            with open(UPLOAD_FOLDER+"/"+filename, "rb") as image_file:
                encoded_img = base64.b64encode(image_file.read()).decode("utf-8") 

            mask_filename = filename.split('.')
            with open(DONE_FOLDER+"/"+mask_filename[0]+'.png', "rb") as image_file:
                encoded_mask = base64.b64encode(image_file.read()).decode("utf-8") 

            return jsonify(
                img=encoded_img,
                mask=encoded_mask
            )

    return '''
    <!doctype html>
    <title>Upload new File</title>
    <h1>Upload new File</h1>
    <form method=post enctype=multipart/form-data>
      <input type=file name=file>
      <input type=submit value=Upload>
    </form>
    '''

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=5001)
