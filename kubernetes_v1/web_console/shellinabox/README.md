# About shellinabox

Shell In A Box implements a web server that can export arbitrary command line tools to a web based terminal emulator. This emulator is accessible to any JavaScript and CSS enabled web browser and does not require any additional browser plugins.

# Usage

## Running command

shellinaboxd --port=8888 --user-css Normal:+white-on-black.css,Reverse:-black-on-white.css --disable-ssl-menu -t -s /:LOGIN

## Access

http://localhost:8888/
