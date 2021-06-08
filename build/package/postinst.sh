#!/bin/bash


# Exit on error
set -e

# Exit if variable is undefined
set -u


if [ -f /usr/bin/openitcockpit-agent ]; then

    if [ -x "$(command -v systemctl)" ]; then
        if [ -d /lib/systemd/system/ ]; then
            # Debian / Ubuntu / Arch
            if [ ! -f /lib/systemd/system/openitcockpit-agent.service ]; then
                ln /etc/openitcockpit-agent/init/openitcockpit-agent.service /lib/systemd/system/openitcockpit-agent.service
            fi
        elif [ -d /usr/lib/systemd/system/ ]; then
            # RedHat / Suse
            if [ ! -f /usr/lib/systemd/system/openitcockpit-agent.service ]; then
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
            if [ -x "$(command -v update-rc.d)" ]; then
                # Debian / Ubuntu
                update-rc.d -f openitcockpit-agent defaults
            fi
            if [ -x "$(command -v chkconfig)" ]; then
                # CentOS
                chkconfig openitcockpit-agent on
            fi
        fi
        
        service openitcockpit-agent start
    fi

fi

if [ -f /Applications/openitcockpit-agent/openitcockpit-agent ]; then
    if [ -f /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist ]; then
        if [ -d /Library/LaunchDaemons/ ] && [ ! -f /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist ]; then
            ln -s /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist
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

    # Keep configs on Updates
    if [ -f /Applications/openitcockpit-agent/config.ini.old ]; then
        cp /Applications/openitcockpit-agent/config.ini.old /Applications/openitcockpit-agent/config.ini
    fi

    if [ -f /Applications/openitcockpit-agent/customchecks.ini.old ]; then
        cp /Applications/openitcockpit-agent/customchecks.ini.old /Applications/openitcockpit-agent/customchecks.ini
    fi

    if [ "$enableConfig" == "1" ]; then
        /bin/launchctl load /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist
    fi
    
    if [ ! -d "/Library/Logs/openitcockpit-agent" ]; then
        mkdir -p /Library/Logs/openitcockpit-agent
    fi

    /bin/launchctl start com.it-novum.openitcockpit.agent

fi