# Dev Tools CLI

This is a very simple tool that allows me to quickly automate tasks that I do often.

## Commands

### Generate

#### `generate key`

Generates a random key of the specified amount of bytes and encodes it with the specified encoding. Hex and Base64
are supported.

#### `generate password`

Generates a random password of the specified length. Includes options for disabling symbols and numbers.
Has support for setting minimum number of symbols and numbers.

#### `generate uuid`

Generates a random UUID V4.

### Encode

#### `encode webp`

Recursively converts all images in the specified directory and its subdirectories to webp format. Outputs them to your
specified output directory.

### Edit

#### `edit rename`

Recursively renames all files in the specified directory and its subdirectories. Currently it only supports
renaming files and directories to be lower case and URL friendly.

## Requirements

If you're using the `encode webp` command, you'll need to have the `cwebp` binary installed. You can get it from
[here](https://developers.google.com/speed/webp/download).