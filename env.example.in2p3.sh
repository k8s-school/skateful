# Parameters used to generate pv/pvc specification

INSTANCE="qserv"

# Use current namespace
NS=$(kubectl config view  --output 'jsonpath={..namespace}')
NS=$([ ! -z "$NS" ] && echo "$NS" || echo "default")

PARALLEL_SSH_CFG="$HOME/.ssh/sshqserv"
PARALLEL_SSH_MASTER="$HOME/.ssh/sshqservmaster"

REPL_DB_HOST="ccqserv201"
INGEST_DB_HOST="ccqserv201"

ROOT_DATA_DIR="/data"
DATA_DIR="${ROOT_DATA_DIR}/${NS}"

MASTERS="ccqserv201 ccqserv202"
WORKERS="ccqserv203 ccqserv204 ccqserv205 ccqserv206 ccqserv207 ccqserv208 ccqserv209 ccqserv210 ccqserv211 ccqserv212 ccqserv213 ccqserv214 ccqserv215 ccqserv216 ccqserv217 ccqserv218 ccqserv219 ccqserv220 ccqserv221 ccqserv222"

SSH_CFG_OPT="-t"
