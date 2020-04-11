#!/bin/bash
set -ex

# Create gcloud project iam account
gcloud iam service-accounts create k8s-tekton-account

# Create storage bucket
gsutil mb gs://zagent-tekton-builds-2020

# Set ACL to allow service account write access to the bucket
gsutil acl ch -u  k8s-tekton-account@doubting-zcash.iam.gserviceaccount.com:WRITE gs://zagent-tekton-builds-2020

# Get a service auth file for the service account
gcloud iam service-accounts keys create k8s-tekton-account.json --iam-account k8s-tekton-account@doubting-zcash.iam.gserviceaccount.com

# Create kubernetes secret from service file
kubectl create secret generic k8s-tekton-account --from-file=k8s-tekton-account.json 
