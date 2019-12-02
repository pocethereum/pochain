#!/bin/bash

echo "start make" \
&& make \
&& echo "start restart poc" \
&& (pkill dlv; pkill poc;sleep 1;  pkill -9 dlv; pkill -9 poc; echo "kill done") \
&& (rm -r /Users/bing/Library/Poc/poc || echo "")\
&& (nohup dlv exec ./build/bin/poc \
            --headless --listen=:2345 --api-version=2 --accept-multiclient \
            --\
            --mine  --gcmode archive \
            --debug --verbosity 5 --vmdebug --rpc --rpcapi web3,net,account,personal,eth,minedev,rpc \
            --ethstats mac:mac@192.168.0.31:3000 &) \
&& echo "start done" \

#
#&& BINGDEV="true" make \
#&& sleep 2.5 \
#&& (pocpid=`ps -ef | grep "./build/bin/poc" | grep -v grep | awk '{print $2}'`; echo "pocpid=$pocpid"; nohup dlv attach $pocpid --headless --listen=:2345 --api-version=2 --accept-multiclient &) \
#&& ./build/bin/poc attach /Users/bing/Library/Poc/poce.ipc
#       ./build/bin/poc --mine  --gcmode archive \
#                   --debug --verbosity 5 --vmdebug --rpc --rpcapi web3,net,account,personal,eth,minedev,rpc \
#                   --pocstats mac:mac@192.168.0.31:3000 2>&1 | tee _all.log |grep INFO &) \
#
# nohup dlv exec ./build/bin/poc --headless --listen=:2345 --api-version=2 --accept-multiclient -- attach http://192.168.0.31:8545
