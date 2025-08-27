#!/bin/bash

set -e

INVENTORY=hosts.ini
PLAYBOOK=bootstrap.yml

ansible-playbook -i $INVENTORY $PLAYBOOK
