# A tool use to translate English in terminal

## Install

First,with `git` installed, clone the repository and run the `install.sh` script.

```sh
git clone https://github.com/cncsmonster/tl-go
cd tl-go
./install.sh
```

Or 

```sh
go install github.com/cncsmonster/tl-go@latest
```

## Usage

i usually alias the command to `tl` in my `.bashrc` or `.zshrc` file.

```sh
alias tl='tl-go'
```

Then you can use it like this:

```sh
tl "hello world"
```