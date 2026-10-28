#!/bin/bash
cd "$(dirname "$0")" && go build -o depviz ./src && ./depviz "$@"
