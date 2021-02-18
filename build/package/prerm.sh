#!/bin/bash


# Exit on error
set -e

# Exit if variable is undefined
set -u


if [ -f /usr/bin/openitcockpit-agent ]; then

    set +e
    if [ -x "$(command -v systemctl)" ]; then
        /bin/systemctl -a | grep openitcockpit-agent >/dev/null
        RC=$?
        if [ "$RC" -eq 0 ]; then
            /bin/systemctl stop openitcockpit-agent
            /bin/systemctl disable openitcockpit-agent
        fi      
    else
        invoke-rc.d openitcockpit-agent stop
        update-rc.d -f openitcockpit-agent remove
        
    fi
    rm -f /etc/init.d/openitcockpit-agent /lib/systemd/system/openitcockpit-agent.service /usr/lib/systemd/system/openitcockpit-agent.service
    set -e

fi

if [ -f /Applications/openitcockpit-agent/openitcockpit-agent ]; then

    touch /Applications/openitcockpit-agent/tmp_runrm

    set +e
    /bin/launchctl list | grep com.it-novum.openitcockpit.agent >/dev/null
    RC=$?
    if [ "$RC" -eq 0 ]; then
        /bin/launchctl stop com.it-novum.openitcockpit.agent
        /bin/launchctl unload -F /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist
    fi
    set -e
    
    rm -rf /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist /Applications/openitcockpit-agent/config.ini /Applications/openitcockpit-agent/customchecks.ini /Applications/openitcockpit-agent /private/etc/openitcockpit-agent
fi
