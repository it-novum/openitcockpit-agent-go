#!/bin/bash


# Exit on error
set -e

# Exit if variable is undefined
set -u


if [ -f /usr/bin/openitcockpit-agent ]; then

    if [ -x "$(command -v systemctl)" ]; then
        if [ ! -f /lib/systemd/system/openitcockpit-agent.service ]; then
            if [ -d /lib/systemd/system/ ]; then
                # Debian / Ubuntu / Arch
                ln /etc/openitcockpit-agent/init/openitcockpit-agent.service /lib/systemd/system/openitcockpit-agent.service
            elif [ -d /usr/lib/systemd/system/ ]; then
                # RedHat / Suse
                ln /etc/openitcockpit-agent/init/openitcockpit-agent.service /usr/lib/systemd/system/openitcockpit-agent.service
            fi
        fi
        
        systemctl daemon-reload
        systemctl start openitcockpit-agent
        systemctl enable openitcockpit-agent
    else
        
        enableConfig="0"
        if [ ! -f /etc/init.d/openitcockpit-agent ]; then
            enableConfig="1"
            ln /etc/openitcockpit-agent/init/openitcockpit-agent.init /etc/init.d/openitcockpit-agent
        fi
        
        if [ "$enableConfig" == "1" ]; then
            update-rc.d -f openitcockpit-agent defaults
        fi
        
        invoke-rc.d openitcockpit-agent start
    fi

fi

if [ -f /Applications/openitcockpit-agent/openitcockpit-agent ]; then
    if [ -f /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist ]; then
        if [ -d /Library/LaunchDaemons/ ]; then
            ln -s /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist /Library/LaunchDaemons/
        fi
    fi

    enableConfig="0"
    set +e
    /bin/launchctl list | grep com.it-novum.openitcockpit.agent
    RC=$?
    if [ "$RC" -eq 1 ]; then
        enableConfig="1"
    fi
    set -e
    
    if [ "$enableConfig" == "1" ]; then
        /bin/launchctl load /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist
    fi
    
    /bin/launchctl start com.it-novum.openitcockpit.agent

fi
