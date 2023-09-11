#!/usr/bin/env sh

die() {
  echo "$1" && exit 1
}

[ -z "$MENTOR_ADD_TOKEN" ] && die "Set MENTOR_ADD_TOKEN envvar"
[ -z "$MENTOR_REMOVE_TOKEN" ] && die "Set MENTOR_REMOVE_TOKEN envvar"
[ -z "$CHECK_ME_TOKEN" ] && die "Set CHECK_ME_TOKEN envvar"
[ -z "$LABS_TOKEN" ] && die "Set LABS_TOKEN envvar"
[ -z "$PRIVATE_CHANNEL_ID" ] && die "Set PRIVATE_CHANNEL_ID envvar"
[ -z "$DEBUG_CHANNEL_ID" ] && die "Set DEBUG_CHANNEL_ID envvar"

sed -i \
  -e "s/MENTOR_ADD_TOKEN/$MENTOR_ADD_TOKEN/g" \
  -e "s/MENTOR_REMOVE_TOKEN/$MENTOR_REMOVE_TOKEN/g" \
  -e "s/CHECK_ME_TOKEN/$CHECK_ME_TOKEN/g" \
  -e "s/MENTOR_LABS_TOKEN/$MENTOR_LABS_TOKEN/g" \
  -e "s/LABS_TOKEN/$LABS_TOKEN/g" \
  -e "s/SET_NAME_TOKEN/$SET_NAME_TOKEN/g" \
  /etc/mostful-manager/config.json

MMST_UID="${MMST_UID:=cbeer_lab}"

/usr/local/bin/mostful-manager \
  -url "$URL" \
  -ownUrl "$OWNURL" \
  -tok "$MMST_TOKEN" \
  -db "$DB_URL" \
  -cfg /etc/mostful-manager/config.json \
  -uid "$MMST_UID" \
  -pchan "$PRIVATE_CHANNEL_ID" \
  -dchan "$DEBUG_CHANNEL_ID"
