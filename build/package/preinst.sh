#!/bin/bash


# Exit on error
set -e

# Exit if variable is undefined
set -u

# Move Agent 1.x configs on Linux
set +e
if [ -f /etc/openitcockpit-agent/config.cnf ]; then
    mv /etc/openitcockpit-agent/config.cnf /etc/openitcockpit-agent/config.ini
fi

if [ -f /etc/openitcockpit-agent/customchecks.cnf ]; then
    mv /etc/openitcockpit-agent/customchecks.cnf /etc/openitcockpit-agent/customchecks.ini
fi
set -e

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

# Move Agent 1.x configs on macOS
set +e
if [ -f /Applications/openitcockpit-agent/config.cnf ]; then
    mv /Applications/openitcockpit-agent/config.cnf /Applications/openitcockpit-agent/config.ini
fi

if [ -f /Applications/openitcockpit-agent/customchecks.cnf ]; then
    mv /Applications/openitcockpit-agent/customchecks.cnf /Applications/openitcockpit-agent/customchecks.ini
fi
set -e


if [ -f /Applications/openitcockpit-agent/openitcockpit-agent ]; then

    if [ -f /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist ]; then
        /bin/launchctl stop com.it-novum.openitcockpit.agent
        /bin/launchctl unload -F /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist
    fi

    # Keep configs on Updates
    if [ -f /Applications/openitcockpit-agent/config.ini ]; then
        cp /Applications/openitcockpit-agent/config.ini /Applications/openitcockpit-agent/config.ini.old
    fi

    if [ -f /Applications/openitcockpit-agent/customchecks.ini ]; then
        cp /Applications/openitcockpit-agent/customchecks.ini /Applications/openitcockpit-agent/customchecks.ini.old
    fi
    
fi
