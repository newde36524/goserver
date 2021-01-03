#! /bin/bash
echo 'restart tcpserver...'
pidof ./tcpserver | xargs kill -9
if [ "$?" = "0" ]; then
nohup ./tcpserver > /dev/null 2>&1 &
if [ "$?" = "0" ]; then
echo 'successfully'
jobs -l
fi
fi
