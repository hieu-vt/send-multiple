#!/usr/bin/bash
numberOfCurl="${NUMBER_OF_CURL:-100}"
timeout="${TIMEOUT:-0.01}"
echo "Start run $numberOfCurl connections"
i=0; while [ $i -le $numberOfCurl ]; do
    i=$((i+1))
    curl -s -X GET -H "Accept:text/event-stream" "http://localhost:8081/streaming/price/ANZ.ASX,BHP.ASX,CBA.ASX,RIO.ASX" > /dev/null  & 2> error.log
    sleep $timeout
done

while true
    do
        echo Keep running
        sleep 15
done