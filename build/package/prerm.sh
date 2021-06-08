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

            if [ "$1" -eq "0" ] || [ "$1" = "purge" ] || [ "$1" = "remove" ] ; then
                # Uninstall on CentOS / Debian / Ubuntu
                /bin/systemctl stop openitcockpit-agent
                /bin/systemctl disable openitcockpit-agent
            if


        fi      
    else
        service openitcockpit-agent stop
        if [ -x "$(command -v update-rc.d)" ]; then
            # Debian / Ubuntu
            update-rc.d -f openitcockpit-agent remove
        fi
        if [ -x "$(command -v chkconfig)" ]; then
            # CentOS
            chkconfig openitcockpit-agent off
            chkconfig --del openitcockpit-agent
        fi
        
    fi
    
    if [ "$1" -eq "0" ]; then
        # Uninstall on CentOS
        rm -f /etc/init.d/openitcockpit-agent /lib/systemd/system/openitcockpit-agent.service /usr/lib/systemd/system/openitcockpit-agent.service /var/log/openitcockpit-agent
    fi

    if [ "$1" = "purge" ] || [ "$1" = "remove" ] ; then
        # Uninstall on Debian / Ubuntu
        rm -f /etc/init.d/openitcockpit-agent /lib/systemd/system/openitcockpit-agent.service /usr/lib/systemd/system/openitcockpit-agent.service /var/log/openitcockpit-agent
    fi

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
    
    if [ -d "/Library/Logs/openitcockpit-agent" ]; then
        rm -rf /Library/Logs/openitcockpit-agent
    fi

    rm -rf /Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist /Library/LaunchDaemons/com.it-novum.openitcockpit.agent.plist /Applications/openitcockpit-agent/config.ini /Applications/openitcockpit-agent/customchecks.ini /Applications/openitcockpit-agent /private/etc/openitcockpit-agent
fi
