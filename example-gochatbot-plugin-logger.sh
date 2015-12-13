#!/bin/sh

# the RPC endpoint is exposed to the plugin by the environment variable:
# GOCHATBOT_RPC_BIND
# move this file to the same directory in which gochatbot binary is compiled in,
# and rename it to a name beginning with "gochatbot-plugin-", e.g.
# "gochatbot-plugin-logger.sh"
# Ensure also it has been enabled with execution bit:
# chmod +x gochatbot-plugin-logger.sh

httpEndPoint="http://$GOCHATBOT_RPC_BIND/pop"
while true;
do
	wget $httpEndPoint -q -O - >> chatlog.txt
	sleep 1;
done;