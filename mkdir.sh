#!/bin/bash

#  Create directory belonging to a given user
#  on all hosts

# @author Fabrice Jammes SLAC/IN2P3

set -euxo pipefail

DIR=$(cd "$(dirname "$0")"; pwd -P)
. "$DIR/env.sh"

REMOTE_DIR="$DATA_DIR/qserv"
echo "Create directory $REMOTE_DIR on all nodes, if not exists"

REMOTE_USER="qserv"

parallel --tag --nonall --slf "$PARALLEL_SSH_CFG" "echo '$PASSWORD' | sudo -S su root sh -c 'mkdir -p $DATA_DIR && sudo chown $REMOTE_USER:$REMOTE_USER $DATA_DIR'"
TARGET_DIR="${DATA_DIR}/replication"
parallel --tag --nonall --slf "$PARALLEL_SSH_MASTER" "echo '$PASSWORD' | sudo -S su qserv sh -c 'mkdir -p $TARGET_DIR'"
TARGET_DIR="${DATA_DIR}/ingest"
parallel --tag --nonall --slf "$PARALLEL_SSH_MASTER" "echo '$PASSWORD' | sudo -S su qserv sh -c 'mkdir -p $TARGET_DIR'"
TARGET_DIR="${DATA_DIR}/qserv"
parallel --tag --nonall --slf "$PARALLEL_SSH_CFG" "echo '$PASSWORD' | sudo -S su qserv sh -c 'mkdir -p $TARGET_DIR'"
