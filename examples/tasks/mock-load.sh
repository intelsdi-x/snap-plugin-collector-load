#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# source the common bash script 
. "${__dir}/../../scripts/common.sh"

# ensure PLUGIN_PATH is set
TMPDIR=${TMPDIR:-"/tmp"}
PLUGIN_PATH=${PLUGIN_PATH:-"${TMPDIR}/snap/plugins"}
mkdir -p $PLUGIN_PATH

_info "downloading plugins"
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snap-plugin-publisher-mock-file && chmod 755 snap-plugin-publisher-mock-file)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/snap/latest_build/linux/x86_64/snap-plugin-processor-passthru && chmod 755 snap-plugin-processor-passthru)
(cd $PLUGIN_PATH && curl -sSO http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-load/latest_build/linux/x86_64/snap-plugin-collector-load && chmod 755 snap-plugin-collector-load)

SNAP_FLAG=0

# this block will wait check if snaptel and snapteld are loaded before the plugins are loaded and the task is started
 for i in `seq 1 10`; do
            _info "try ${i}"
             if [[ -f /usr/local/bin/snaptel && -f /usr/local/sbin/snapteld ]];
                then
                    _info "loading plugins"
                    snaptel plugin load "${PLUGIN_PATH}/snap-plugin-publisher-mock-file"
                    snaptel plugin load "${PLUGIN_PATH}/snap-plugin-processor-passthru"
                    snaptel plugin load "${PLUGIN_PATH}/snap-plugin-collector-load"

                    _info "creating and starting a task"
                    snaptel task create -t "${__dir}/task-load.json"

                    SNAP_FLAG=1

                    break
             fi 
        
        _info "snaptel and/or snapteld are unavailable, sleeping for 5 seconds"
        sleep 5
done 


# check if snaptel/snapteld have loaded
if [ $SNAP_FLAG -eq 0 ]
    then
     echo "Could not load snaptel or snapteld"
     exit 1
fi