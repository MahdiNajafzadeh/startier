#!/bin/bash
clear
echo "RUN NODE ID = $EASYNODE_ID"
air -c .air.node.$EASYNODE_ID.toml
