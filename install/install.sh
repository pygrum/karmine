#!/bin/bash
set -e

if [ $# -ne 1 ]
then
    echo "usage: install.sh <domain-name>"
    exit 1
fi

SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

cd ${SCRIPTPATH}/..

CONFIG_LINES=(
    "{"
    "\"bin_path\":\"${HOME}/.kbin\","
    "\"src_path\":\"$(realpath .)\","
    "\"cert_pem\":\"${HOME}/.kdots/karmine.crt\","
    "\"key_pem\":\"${HOME}/.kdots/karmine.key\","
    "\"db\":\"${HOME}/.kdots/karmine.db\","
    "\"ssldomain\":\"${1}\","
    "\"endpoint\":\"/api\""
    "}"
)

go build 

if [ -f "${HOME}/.konfig" ]
then
    echo ".konfig found at ${HOME}. skipping"
    exit 1
fi

for LINE in ${CONFIG_LINES[@]}
do
    echo "$LINE" >> ${HOME}/.konfig
done

if [ -d "${HOME}/.kbin" ]
then
    echo ".kbin already present. exiting"
    exit 1
fi

if [ ! -d "${HOME}/.kdots" ]
then
    mkdir ${HOME}/.kdots
    # OpenSSL >= 1.1.1
    sudo openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
    -keyout ${HOME}/.kdots/karmine.key -out ${HOME}/.kdots/karmine.crt -subj "/CN=${1}" \
    -addext "subjectAltName=DNS:${1}"

    sudo chmod 444 ${HOME}/.kdots/karmine.key
    cp karmine.db ${HOME}/.kdots
fi

./cmd/build.sh

echo "Build complete"