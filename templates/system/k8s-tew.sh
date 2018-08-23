#!/bin/sh

export K8S_TEW_BASE_DIRECTORY={{.BaseDirectory}}

eval $({{.Binary}} environment)
