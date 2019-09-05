#!/bin/bash
yum install https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm -y
yum install Xvfb -y
yum install xorg-x11-fonts* -y
echo "#!/bin/bash" > /usr/bin/xvfb-chrome
echo '_kill_procs() {' >> /usr/bin/xvfb-chrome
echo '  kill -TERM $chrome' >> /usr/bin/xvfb-chrome
echo '  wait $chrome' >> /usr/bin/xvfb-chrome
echo '  kill -TERM $xvfb' >> /usr/bin/xvfb-chrome
echo '}' >> /usr/bin/xvfb-chrome
echo 'trap _kill_procs SIGTERM' >> /usr/bin/xvfb-chrome
echo 'XVFB_WHD=${XVFB_WHD:-1280x720x16}' >> /usr/bin/xvfb-chrome
echo 'Xvfb :99 -ac -screen 0 $XVFB_WHD -nolisten tcp &' >> /usr/bin/xvfb-chrome
echo 'xvfb=$!' >> /usr/bin/xvfb-chrome
echo 'export DISPLAY=:99 ' >> /usr/bin/xvfb-chrome
echo 'chrome --no-sandbox --disable-gpu$@ & ' >> /usr/bin/xvfb-chrome
echo 'chrome=$!' >> /usr/bin/xvfb-chrome
echo 'wait $chrome' >> /usr/bin/xvfb-chrome
echo 'wait $xvfb' >> /usr/bin/xvfb-chrome
chmod +x /usr/bin/xvfb-chrome
ln -s /etc/alternatives/google-chrome /usr/bin/chrome
rm -rf /usr/bin/google-chrome
ln -s /usr/bin/xvfb-chrome /usr/bin/google-chrome
ll /usr/bin/ | grep chrom
