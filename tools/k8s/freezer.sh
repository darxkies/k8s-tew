#!/bin/bash

BINARY="$1"
TARGET_PATH="$2"
BINARY_PATH=$TARGET_PATH/bin
LIBRARY_PATH=$TARGET_PATH/lib

mkdir -p $BINARY_PATH $LIBRARY_PATH

cp "$BINARY" "$BINARY_PATH/$(basename $BINARY)"

for library in `ldd "$BINARY" | cut -d '>' -f2 | awk '{print $1}'`; do
  if [ -f "$library" ] ; then
    cp -v "$library" "$LIBRARY_PATH/$(basename $library)"
  fi  
done

