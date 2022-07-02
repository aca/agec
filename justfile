test:
    go test -v ./...

dev:
    go install .
    agec completion zsh > ~/.zsh/zsh-completions/src/_agec
    agec completion bash | sudo tee /usr/share/bash-completion/completions/agec >/dev/null || true
    agec completion bash | sudo tee /usr/local/share/bash-completion/completions/agec >/dev/null || true
    agec completion fish > ~/.config/fish/completions/agec.fish
