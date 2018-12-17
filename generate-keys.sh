#!/bin/bash

sshdir=${1:-ssh}
comment=${2:-wolserver}

mkdir -p $sshdir
for keytype in rsa dsa ecdsa; do
    sshkey=$sshdir/ssh_host_${keytype}_key
    if [[ ! -f $sshkey ]]; then
        echo "Creating $sshkey"
        ssh-keygen -t $keytype -C $comment -N '' -f $sshkey
        echo ''
    fi
done
