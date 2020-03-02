#!/bin/sh

export K8S_TEW_BASE_DIRECTORY={{.BaseDirectory}}

source <({{.Binary}} environment)
