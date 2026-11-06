# Security Policy

## Supported versions

Only the latest release is supported. Fixes land on `main` and ship in the next release; older tags do not get backports.

## Reporting a vulnerability

Do not open a public issue for a security problem.

Report it privately through GitHub: go to the [Security tab](https://github.com/tapiaw38/edwin/security/advisories/new) and open a draft security advisory. If that is not available to you, email tapiaw38@gmail.com instead.

Please include what the issue is, how to reproduce it, which version or commit you tested, and your distribution and compositor.

You can expect an acknowledgement within a week. Once a fix is ready, it is released and the advisory is published with credit to you, unless you prefer otherwise.

## Scope

Edwin runs as a local desktop application under the user's own account. It launches other programs, reads and writes files, and shells out to optional integrations such as `playerctl`, `yt-dlp`, and `wl-clipboard`. Issues worth reporting include command injection through a query, prefix, or config value; escaping the user's own privileges; and leaking secrets handled by the `pass` integration or the clipboard.

A feature behaving as documented is not a vulnerability — Edwin is a launcher, so running commands the user types is the point.
