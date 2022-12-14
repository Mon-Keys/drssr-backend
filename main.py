import os
from flask import Flask, flash, request, redirect, url_for, send_from_directory, jsonify
from werkzeug.utils import secure_filename
from cut import cut

DIR = os.getcwd()
UPLOAD_FOLDER = DIR + '/uploads'
DONE_FOLDER = DIR + '/cut'
ALLOWED_EXTENSIONS = {'txt', 'pdf', 'png', 'jpg', 'jpeg', 'gif', 'webp'}

app = Flask(__name__)
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
app.config['DONE_FOLDER'] = DONE_FOLDER


@app.route("/", methods=['GET'])
def index():
    return "Hello, "


@app.route("/about", methods=['GET'])
def index1():
    return "about, "


def allowed_file(filename):
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


@app.route('/upload', methods=['GET', 'POST'])
def upload_file():
    if request.method == 'POST':
        if 'file' not in request.files:
            flash('No file part')
            return redirect(request.url)
        print("--------1")
        file = request.files['file']
        print("--------2")
        if file.filename == '':
            flash('No selected file')
            return redirect(request.url)
        print("--------3")

        if file and allowed_file(file.filename):
            filename = secure_filename(file.filename)
            file.save(os.path.join(app.config['UPLOAD_FOLDER'], filename))

            # cut_img(filename)  # Магия
            fname = cut(filename)  # Super Магия

            return jsonify(path=fname)
            # return redirect(url_for('uploaded_file',
            #                         filename=fname))
    return '''
    <!doctype html>
    <title>Upload new File</title>
    <h1>Upload new File</h1>
    <form method=post enctype=multipart/form-data>
      <input type=file name=file>
      <input type=submit value=Upload>
    </form>
    '''


@app.route('/uploads/<filename>')
def uploaded_file(filename):
    return send_from_directory(app.config['DONE_FOLDER'], filename)


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=5000)
