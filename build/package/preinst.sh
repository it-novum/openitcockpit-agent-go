#!/bin/bash


# Exit on error
set -e

# Exit if variable is undefined
set -u

if [ -f /usr/bin/openitcockpit-agent ]; then

    if [ -x "$(command -v systemctl)" ]; then
        set +e
        systemctl is-active --quiet openitcockpit-agent
        if [ $? = 0 ]; then
            systemctl stop openitcockpit-agent
        fi
        systemctl disable openitcockpit-agent
        set -e
    else
        set +e
        ps auxw | grep -P '/usr/bin/openitcockpit-agent' | grep -v grep >/dev/null
        if [ $? = 0 ]; then
            invoke-rc.d openitcockpit-agent stop
        fi
        update-rc.d -f openitcockpit-agent remove
        set -e
    fi

fi

if [ -f /Applications/openitcockpit-agent/openitcockpit-agent ]; then

    if [ -f /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist ]; then
        /bin/launchctl stop com.it-novum.openitcockpit.agent
        /bin/launchctl unload -F /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist
    fi
    
fi
