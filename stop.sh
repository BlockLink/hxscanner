#!/bin/bash
HXSCANNER_PID=`ps -aux | grep "hxscanner -db_pass=12345ssdlh" | grep -v grep | awk '{print $2}'`
kill $HXSCANNER_PID

