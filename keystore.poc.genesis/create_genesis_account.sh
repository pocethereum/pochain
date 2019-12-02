#!/bin/bash

for i in `seq 1 10` 
do
    password=$RANDOM$RANDOM$RANDOM
    echo $i"--------$password--"
    echo -e $password"\n"$password"\n" | ../build/bin/poc account new --datadir . 
done
