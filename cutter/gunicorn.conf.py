import multiprocessing

bind = "localhost:5032"
workers = 4
threads = 6
accesslog = "/var/tmp/gunicorn/access.log"
