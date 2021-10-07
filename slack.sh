#!/usr/bin/env bash

# add your webhook
# example https://hooks.slack.com/services/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
YourWebhook=
curl -X POST -H 'Content-type: application/json' --data '{"text":"pod metadata changed"}' $YourWebhook