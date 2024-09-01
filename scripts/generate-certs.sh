#!/bin/bash

# ATTENTION:
## I HAVE HARD-CODED THE FOLLOWING CA CONFIG FILES:
## ca-controllers.conf
## ca-nodes.conf
## (see files folder)

set -euo pipefail

check_openssl_version() {
    min_o_version=3
    o_version=$(openssl version  | awk {'print $2'} | awk -F "." {'print $1'})
    if [[ ${o_version} < ${min_o_version} ]]; then
        echo "[$(date +%s)] [FAILED] - openssl version is too old, upgrade it first."
        exit 1;
    fi 
}

gen_cert() {
    certs_dir=$1
    config_file=$2
    cid=$3

    echo "[$(date +%s)] [INFO] - generating key for $cid..."
    openssl genrsa -out "${certs_dir}/${cid}.key" 4096

    echo "[$(date +%s)] [INFO] - generating csr for $cid..."
    openssl req -new -key "${certs_dir}/${cid}.key" -sha256 -config "${config_file}" -section ${cid} -out "${certs_dir}/${cid}.csr"

    echo "[$(date +%s)] [INFO] - generating cert for $cid..."
    openssl x509 -req -days 3653 -in "${certs_dir}/${cid}.csr" -copy_extensions copyall -sha256 -CA "${certs_dir}/ca.crt" -CAkey "${certs_dir}/ca.key" -CAcreateserial -out "${certs_dir}/${cid}.crt"
}

main() {
    certs_dir="$1"
    config_file="$2"
    cids="${@: 3}"

    mkdir ${certs_dir} || true

    for cid in ${cids[@]}; do
        if [[ ${cid} == "ca" ]]; then
        #TODO: IF CA is regenerated, regen everything else too
            echo "[$(date +%s)] [INFO] - generating ca configs"
            openssl genrsa -out "${certs_dir}/${cid}.key" 4096
            openssl req -x509 -new -sha512 -noenc -key "${certs_dir}/${cid}.key" -days 3653 -config "${config_file}" -out "${certs_dir}/${cid}.crt"
        else
            echo "[$(date +%s)] [INFO] - generating tls config for ${cid}"
            gen_cert $certs_dir $config_file $cid
        fi
    done
}



check_openssl_version
main $@
