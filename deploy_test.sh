#!/bin/bash
#####################################################################
###############deploy helper ################ begin #################
#####################################################################
POCHOST=gateway.poc.com

echo "start make" \
&& make poc-linux-amd64 \
&& echo "mv remote poc to poc.bak" \
&& (ssh pocethereum@$POCHOST 'cd /home/pocethereum/poc/bin && mv poc poc.bak;' || echo "backup error") \
&& echo "start scp poc" \
&& scp ./build/bin/poc-linux-amd64 pocethereum@$POCHOST:/home/pocethereum/poc/bin/poc \
&& echo "start reset poc" \
&& ssh pocethereum@$POCHOST 'cd /home/pocethereum/poc&& ./load.sh restart;' \
&& echo "done" \
&& echo "done" \
&& echo "done" 

######## reset all data ##############################################
#&& ssh pocethereum@$POCHOST 'cd /home/pocethereum/ && sh reset.sh;'\
######## reset all data ##############################################
