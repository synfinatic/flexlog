# flexlog

The Flexible slog.Handler

## About

I wanted to switch from logrus to slog in [aws-sso-cli](https://github.com/synfinatic/aws-sso-cli),
but I found the default slog library a bit too basic for my tastes.  Additionally, one feature I
really <3 about logrus was the ability to write unit tests against generated logs; a feature which
is currently missing from slog.

Anyways, I decided to breakout my work into a separate library for future use.  Maybe it'll work
for you too?