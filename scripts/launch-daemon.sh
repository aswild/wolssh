#!/bin/bash

THISDIR="$(readlink -f "$(dirname "$0")")"
DAEMON="${THISDIR}/wolssh-mips64"
DAEMON_ARGS=""

LOGFILE="${THISDIR}/wolssh.log"

echo "DAEMON = $DAEMON"

/sbin/start-stop-daemon --start --quiet --make-pidfile --pidfile wolssh.pid --background \
    --startas /bin/bash -- -c "cd '${THISDIR}' && exec '$DAEMON' $DAEMON_ARGS >>${LOGFILE} 2>&1"
