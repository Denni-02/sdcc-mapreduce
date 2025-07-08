#!/bin/bash

cat > .env <<EOF
#Inserisci qui le credenziali temporanee dal Learner Lab

ENABLE_S3=true
S3_BUCKET=sdcc-mapreduce-recovery

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_SESSION_TOKEN=
EOF

echo ".env creato nella root del progetto. Inserisci le credenziali AWS temporanee nel file."
