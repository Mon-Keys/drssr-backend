#!/bin/sh

yes | sudo cp /home/ubuntu/backend/nginx.conf /etc/nginx/sites-enabled/default
sudo systemctl restart nginx >/dev/null 2>&1 &