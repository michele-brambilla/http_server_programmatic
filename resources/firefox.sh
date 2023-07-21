#!/bin/bash
pgrep firefox >& /dev/null && echo 1 || echo 0
