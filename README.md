# dedupe-linker

Dedupe regular files on a filesystem by hard linking and a creating a content-addressible look-aside.

This defaults to `sha1` checksums, but you can specify others like `sha256` or `sha512`.

This also means that the comparison on whether something is a duplicate is based on checksum, and does not consider mode, owner, or xattrs.

This is really a pet-project, but changes welcome.
And is only focused on Linux compatability, but changes welcome.

## Install

```shell
go install github.com/vbatts/dedupe-linker@latest
```

## Usage

The default look-aside base directory is in `~/.dedupe-linker`, so the following would dedupe your home directory:

```shell
dedupe-linker -w $(nproc) ~/
```

Cleaning up from the look-aside needs only to check the links as a ref-count:

```shell
find ~/.dedupe-linker/ -type f -links 1
```

Then delete these files that only have 1 reference:

```shell
find ~/.dedupe-linker/ -type f -links 1 -exec rm -f "{}" \;
```

