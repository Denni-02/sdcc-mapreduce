#!/bin/bash

BUCKET="sdcc-mapreduce-output" 
FILE="output/final_output.txt"
REGION="eu-east-1"            

if [ -f "$FILE" ]; then
  echo "Uploading $FILE to s3://$BUCKET/"
  aws s3 cp "$FILE" "s3://$BUCKET/" --region "$REGION" --acl public-read
else
  echo "File $FILE not found"
fi
