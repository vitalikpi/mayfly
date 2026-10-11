# Mayfly

Ridiculously simple yet functional secret sharing service.

## What does it do?

1. Share your dirty little secret in the text box.
2. Click share
3. You get a link. Server stores your secret ciphertext.
4. Share the link with your friend.
5. The link can only be opened once.
6. Your friend has 24 hours before it's gone forever.

## What's under the hood?

200 lines of go, plain JavaScript, and a sprinkle of HTML with a grain of CSS.

No dependencies, no build chain attacks, no supply chain attacks - there is nothing, just a compiler.

## I don't trust it

You should not. Open your damn dev tools and see what goes in and what goes out. 

## BTW

Oh and by the way - secrets do not survive a restart. Sorry, but it's much better than having them on a disk.

## Self hosting?

Grab yourself systemd, build and copy that single executable.

Put Mayfly behind a reverse-proxy with HTTPS and you are golden.

## AI Vibe?

Partially. But it does not matter. It's 250 lines of code. Read them like in the good old days.
I bet you read larger PRs at work.