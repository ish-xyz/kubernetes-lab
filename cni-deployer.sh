#!/bin/bash
kubectl --kubeconfig /etc/kubernetes/admin.kubeconfig apply -f /etc/kubernetes/default-roles.yaml
helm repo add cilium https://helm.cilium.io/
helm repo update
helm --kubeconfig /etc/kubernetes/admin.kubeconfig install cilium cilium/cilium \
  --version 1.16.1 \
  --namespace kube-system \
  --set kubeProxyReplacement="true" \
  --set k8sServiceHost=kube-apiserver-mytestcluster.compute.zone \
  --set k8sServicePort=6443


#scp -i /home/waffle34/.ssh/capi-demo -o StrictHostKeyChecking=no /home/waffle34/repos/kubernetes-lab/bootstrap-manager/bootstrap-manager zero@54.170.17.190:/home/zero/bootstrap-manager
#