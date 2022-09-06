#!/bin/sh

sed -i '' \
  -e "s/MENTOR_ADD_TOKEN/$MENTOR_ADD_TOKEN/g"\
  -e "s/MENTOR_REMOVE_TOKEN/$MENTOR_REMOVE_TOKEN/g"\
  -e "s/CHECK_ME_TOKEN/$CHECK_ME_TOKEN/g"\
  -e "s/LABS_TOKEN/$LABS_TOKEN/g"\
  /etc/mostful-manager/config.json

/usr/local/bin/mostful-manager -url $URL -tok $MMST_TOKEN -db $DB_URL -cfg /etc/mostful-manager/config.json
