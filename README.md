# bskypost

A quick hack of a golang command line tool to post to Bluesky.

You'll need an application key/password. You can generate one by logging into
Bluesky and going to your Settings, Privacy and Security and App passwords.

Usage:
```
make
export BSKY_USERNAME=myusername.bsky.social
export BSKY_APPKEY=my12-app3-key4-5678
echo "This is my test message" | ./bskypost
```

The tool supports converting URLs in the message into links, but does not yet
support mentions or other facets.

