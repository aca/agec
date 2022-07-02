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
curl -s -L "https://github.com/aca/agec/releases/download/v1.0.0/agec_1.0.0_linux_amd64.tar.gz" | tar xvz agec
sudo mv agec /usr/local/bin
```

Darwin
```
curl -s -L "https://github.com/aca/agec/releases/download/v1.0.0/agec_1.0.0_darwin_all.tar.gz" | tar xvz agec
sudo mv agec /usr/local/bin
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

Clone repository, `examples/` will be the root directory to test agec.
Or just start from any directory with `agec init`.
```
git clone https://github.com/aca/agec.git
cd agec/examples
```

Check configuration already created
```
cat .agec.yaml
```

Add yourself as a user and member of existing group `admin`, with public keys from github
```
curl -s "https://github.com/aca.keys" | agec useradd aca -g admin -R -
```

Create encrypted file that can be decrypted by only "aca" or members of group `admin`
```
agec encrypt secret.txt -u aca -g admin
```

decrypt file, it will try to decrypt file with keys in ~/.ssh by default.
```
agec decrypt secret.txt.age
```

edit files

change owner of file to member james
```
agec chown -u james -g '' secret.txt
```

Re-encrypt it, but you won't be able to decrypt the secret
```
agec encrypt secret.txt
agec decrypt secret.txt.age # fail
```

Try to decrypt it with james's private key
```
agec decrypt secret.txt.age -i james.agekey # success
```

List of available commands, and detailed usage.
```
agec --help
agec [command] --help
```
