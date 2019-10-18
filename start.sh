#!/bin/bash
nohup ./hxscanner -db_pass=12345ssdlh -db_port=18030 -node_endpoint=ws://127.0.0.1:8090 > /dev/null 2>&1 &

