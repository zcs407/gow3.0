#!/bin/bash
pip install selenium
unzip gow-bo/chromedriver_linux64.zip
cp chromedriver /usr/sbin/
chmod +x /usr/sbin/chromedriver
pip install beautifulsoup4
