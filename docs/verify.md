# Checking a download

Every file on the [releases page](https://github.com/confuserdotcom/hourdial/releases)
is built by GitHub Actions from the tagged source, never on a laptop. Two
checks, from weakest to strongest. Run them in the folder you downloaded to.

## 1. The file is the one that was published

```sh
shasum -a 256 -c SHA256SUMS --ignore-missing
```

For the `.dmg`, use its own list: `shasum -a 256 -c tock-<version>.dmg.sha256`.

`OK` next to your file means it matches the list published with the
release. This catches a broken or swapped download, not a swapped list.

## 2. GitHub built it, from this repository, at that tag

With the [GitHub CLI](https://cli.github.com):

```sh
gh attestation verify tock-<version>-linux-amd64.tar.gz -R confuserdotcom/hourdial
```

It checks a signed record (Sigstore) that the release workflow made the
file from a commit in confuserdotcom/hourdial; the output names the workflow
and the tag. It works for the tarballs, the `.dmg`, `SHA256SUMS`, the SBOM
and `tock.rb`, from v1.0.0 on. Verify `SHA256SUMS` this way and step 1
becomes as strong as step 2 for every file it lists.

## The built-in updater

hourdial never downloads itself from GitHub, and no command replaces the
installed binaries with a release asset. Releases are checked by you,
with the two steps above.

The only thing that installs a build by itself is the Linux updater
(`[sync] update = true`, off by default): trackd on a Linux machine
installs the build your Mac publishes to your own sync folder
(`internal/update`). The sums it checks come from that same folder, so
they are not trusted on their own. What it trusts is an ed25519 key:

- The Mac signs `manifest.json` (version, build time, arch and the sha256
  of `tock` and `trackd`) with a private key that never enters the sync
  folder (`tockpub` refuses one inside it).
- The Linux machine takes the public key only through `tock update trust
  <fingerprint>`, with the fingerprint you read on the Mac's own terminal
  (`make bootstrap-hint`). A key in the sync folder is never trusted by
  itself, and a different one there stops updates until you trust it.
- Before anything is installed, trackd checks the signature, that the
  manifest names this machine's arch, that the build is newer than the
  one running, and the sha256 of the bytes it copies; any failure leaves
  the old binaries in place.

So someone who can write to the sync folder, but not to your Mac's key,
can stop updates (by deleting or breaking files) but cannot make the
Linux machine run their code, an older build, or the other arch's build.
A build without a build time baked in (`go build` by hand, not `make`)
takes any signed build, older ones included.

## What is inside

`tock-<version>.cdx.json` is the SBOM: every Go module the build uses, with
its version and hash, in CycloneDX. Feed it to a scanner
(`grype sbom:tock-<version>.cdx.json`, `osv-scanner --sbom=…`) to check for
known vulnerabilities after the release. Dependencies are also checked
with `govulncheck` before every release.

## macOS

The app is signed ad hoc, not notarized: hourdial has no Apple Developer ID.
macOS will say it cannot check the app for malicious software. The checks
above are what stands in for notarization: run them on the `.dmg`, then
System Settings → Privacy & Security → Open Anyway. `tock setup` then signs
what it installs with a certificate it makes on your Mac, so the
Accessibility grant survives upgrades.

Homebrew builds from the source tarball, whose sha256 is in `tock.rb`.

## Building it yourself

The surest check is not to download a binary at all:

```sh
git clone https://github.com/confuserdotcom/hourdial && cd hourdial
git checkout <version>   # a release tag, e.g. the newest on the releases page
make install          # macOS; Linux: make install-linux
```
