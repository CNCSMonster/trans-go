# A tool use to translate English in terminal

## Install

First,with `git` installed, clone the repository and run the `install.sh` script.

```sh
git clone https://github.com/cncsmonster/trans-go
cd trans-go
./install.sh
```

Or 

```sh
go install github.com/cncsmonster/trans-go@latest
```

## Usage

i usually alias the command to `trans` in my `.bashrc` or `.zshrc` file.

```sh
alias trans='trans-go'
```

Then you can use it like this:

```sh
trans "hello world"
```
