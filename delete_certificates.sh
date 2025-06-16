#!/bin/bash

# AWS CLI profile (optional, defaults to default profile)
# export AWS_PROFILE="your-profile-name"

# AWS region (optional, defaults to configured region)
# export AWS_DEFAULT_REGION="your-aws-region"

echo "Starting AWS IoT Core certificate deletion process..."

# Get a list of all IoT certificates
CERTIFICATES=$(aws iot list-certificates --query 'certificates[*].[certificateArn,certificateId]' --output text)

if [ -z "$CERTIFICATES" ]; then
  echo "No IoT Core certificates found to delete."
  exit 0
fi

echo "Found the following certificates:"
echo "$CERTIFICATES"

echo "Processing each certificate..."

while IFS=$'\t' read -r CERT_ARN CERT_ID; do
  echo "----------------------------------------------------"
  echo "Processing Certificate ID: $CERT_ID (ARN: $CERT_ARN)"

  # 1. List and detach policies attached to the certificate
  echo "  Listing and detaching policies..."
  POLICIES=$(aws iot list-attached-policies --target "$CERT_ARN" --query 'policies[*].policyName' --output text)

  if [ -n "$POLICIES" ]; then
    for POLICY_NAME in $POLICIES; do
      echo "    Detaching policy: $POLICY_NAME from $CERT_ID"
      aws iot detach-policy --policy-name "$POLICY_NAME" --target "$CERT_ARN"
      if [ $? -eq 0 ]; then
        echo "    Successfully detached policy: $POLICY_NAME"
      else
        echo "    Failed to detach policy: $POLICY_NAME. Skipping deletion of this certificate."
        continue 2 # Skip to the next certificate
      fi
    done
  else
    echo "    No policies attached to this certificate."
  fi

  # 3. Deactivate the certificate
  echo "  Deactivating certificate: $CERT_ID..."
  aws iot update-certificate --certificate-id "$CERT_ID" --new-status INACTIVE
  if [ $? -eq 0 ]; then
    echo "  Successfully deactivated certificate: $CERT_ID"
  else
    echo "  Failed to deactivate certificate: $CERT_ID. Skipping deletion."
    continue # Skip to the next certificate
  fi

  # Add a small delay to ensure propagation
  sleep 1

  # 4. Delete the certificate
  echo "  Deleting certificate: $CERT_ID..."
  # Using --force-delete is generally not recommended unless you are absolutely sure,
  # as it bypasses the thing attachment check. We've already handled policies and deactivation.
  aws iot delete-certificate --certificate-id "$CERT_ID" --force-delete
  if [ $? -eq 0 ]; then
    echo "  Successfully deleted certificate: $CERT_ID"
  else
    echo "  Failed to delete certificate: $CERT_ID."
  fi

done <<< "$CERTIFICATES"

echo "----------------------------------------------------"
echo "AWS IoT Core certificate deletion process completed."