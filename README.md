### Mostful manager
## What is this?

A spiritual continuation for Slacky Manager, which is not currently in development for "reasons"

This is a bot for Mattermost, which can somewhat distribute work from many requesters to many 
workers. We use it to check laboratory works of students in ISCRA, however it can easily be 
modified to distribute other types of work.

## How do I use this?

Binary produced accepts the following arguments:
- `-url` - The URL at which your wanted Mattermost instance is exposed
- `-tok` - Bot token, see the Mattermost documentation on how you can get one
- `-db` - URL to the database, where you would want to store your precious data
  (currently only PostgreSQL is supported, thus the expected format is 
   "postgres://dbuser@dbaddress:dbport/dbname", you can play with that)
- `-cfg` - config file path. See what's inside it at config.json

For one's convenience, it's possible to containerize it with Docker. There's no funny business 
with building the image, `docker build -t whatevertag .` is absolutely fine.\
To start the image be sure to set the following envvars, these are used for configuring the daemon:
- `MENTOR_ADD_TOKEN` -  see config.json
- `MENTOR_REMOVE_TOKEN` - see config.json
- `CHECK_ME_TOKEN` - see config.json
- `LABS_TOKEN` - see config.json
- `URL` - see -url flag
- `MMST_TOKEN` - see -tok flag
- `DB_URL` - see -db flag

## I think stuff's broken...

Report an issue! This is the best way for me to not forget and eventually make the needed
changes

## Hey, stuff's fun! Can I have a bite?

Sure, but make sure you don't break the licence. I'm no lawyer, but pretty sure you'll be fine
as long as you don't market your changes without my notice :)

## ISCRA's cool and good, but I can do better!

Contributions are of course welcome. With enough sanity your changes could be the future!
There's no COC, maybe there will never be one. As soon as the need arises, I'll make sure it's
known