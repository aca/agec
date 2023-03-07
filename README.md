# agec

**age** en**c**rypt. Yet another attempt to store, manage and share secrets in git repository based on [age](https://github.com/FiloSottile/age).

## Background

There's plenty of mature solutions for this, [sops](https://github.com/mozilla/sops), [git-crypt](https://github.com/AGWA/git-crypt), [blackbox](https://github.com/StackExchange/blackbox), [agebox](https://github.com/slok/agebox), [git-agecrypt](https://github.com/vlaci/git-agecrypt).
I was frustrated with the way it all worked. I wanted something with

- Simple workflow, simple encryption with just SSH keys
- Better shell experience
  - Shell completions (bash, zsh, fish)
  - Invoke command from any subdirectory
- Mechanism to share secrets to limited users/groups in repository.

agec is basically just a small wrapper around [age](https://github.com/FiloSottile/age).

## Installation

Download binary from [releases](https://github.com/aca/agec/releases)

Linux
```
curl -L -o agec "https://github.com/aca/agec/releases/download/v0.2.0/agec_0.2.0_linux_amd64"
chmod +x ./agec
sudo mv ./agec /usr/local/bin
```

Darwin
```
curl -L -o agec "https://github.com/aca/agec/releases/download/v0.2.0/agec_0.2.0_darwin_all"
chmod +x ./agec
sudo mv ./agec /usr/local/bin
```

or build from source, agec requires go >= 1.18
```
go install github.com/aca/agec@main
```

Shell completions require additional setup, supports bash/zsh/fish
```
agec completion [SHELL] --help
```

## Example workflow
Change "aca" with your github id. This example will use public keys registered in github for encryption.

Setup test directory
```
mkdir testdir && cd testdir && git init && agec init && echo "secret txt" > secret.txt
```

Add group "admin" and register "aca" and yourself as a member of group `admin`, with public keys from github
```
agec groupadd admin
curl "https://github.com/aca.keys" | agec useradd aca -g admin -R -
curl "https://github.com/{{ your github id }}.keys" | agec useradd {{ your github id }} -g admin -R -
```

Agec have concept of 'user', 'group'. You can check it in root configuration.
```
cat .agec.yaml
```

Create encrypted file that can be decrypted by members of group `admin`
```
agec encrypt secret.txt -g admin
```

decrypt file, it will try to decrypt file with keys in ~/.ssh by default.
```
agec decrypt secret.txt.age
```

edit files

chown updates owner of the secret, this will change owner of secret.txt from "group:admin" to "user:aca"
```
agec chown -u aca -g '' secret.txt
```

Re-encrypt it, but you won't be able to decrypt the secret as you are not the owner of secret anymore.
```
agec encrypt secret.txt
agec decrypt secret.txt.age # fail
```

List of available commands, and detailed usage.
```
agec --help
agec [command] --help
```
