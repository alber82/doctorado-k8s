#!/bin/bash


/manager \
  --leader-elect \
  --watch-namespaces=${WATCH_NAMESPACES:-''}

