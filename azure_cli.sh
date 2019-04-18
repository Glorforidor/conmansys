#!/bin/bash

set -eu -o pipefail

install () {
	rpm --import https://packages.microsoft.com/keys/microsoft.asc
	zypper addrepo \
        --name 'Azure CLI' \
        --check https://packages.microsoft.com/yumrepos/azure-cli azure-cli
	zypper install --from azure-cli -y azure-cli
}

uninstall () {
	zypper remove -y azure-cli
	zypper removerepo azure-cli
    MSFT_KEY=$(rpm -qa gpg-pubkey /* \
        --qf "%{version}-%{release} %{summary}\n" \
        | grep Microsoft \
        | awk '{print $1}')
	rpm -e --allmatches gpg-pubkey-${MSFT_KEY}
}

case $1 in
    install)
        install
        ;;
    uinstall)
        uninstall
        ;;
    *)
        echo $"Usage: $0 {install|uinstall}"
        exit 1
        ;;
esac
