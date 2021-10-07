#!/bin/bash
echo "CONF:"
cat conf.json
echo -e "\nLOG:"
for i in {1..10};
do
    echo $i
    time go run main.go;
    sleep 20
done
