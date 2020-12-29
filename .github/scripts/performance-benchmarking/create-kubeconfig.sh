#!/bin/bash

set -eo pipefail

if [[ -z "$CONFIG" || -z "$ROLE" || -z "$DIR" ]]; then
    echo "Usage: CONFIG=\"\" ROLE=\"\" DIR=\"\" ./create-kubeconfig.sh"
fi

echo "
apiVersion: v1
clusters:
- cluster:
  "$CONFIG"
contexts:
- context:
    cluster: eks-cluster-1
    user: github-actions-dolt
  name: github-actions-dolt-eks-cluster-1
current-context: github-actions-dolt-eks-cluster-1
kind: Config
preferences: {}
users:
- name: github-actions-dolt
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws-iam-authenticator
      args:
        - \"token\"
        - \"-i\"
        - \"eks-cluster-1\"
        - \"-r\"
        -  \"$ROLE\"
" > "$DIR"/kubeconfig

