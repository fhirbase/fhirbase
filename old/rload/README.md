```
curl https://sh.rustup.rs -sSf | sh\n
echo 'export PATH="$HOME/.cargo/bin:$PATH"'  >> ~/.zshrc
source ~/.zshrc
rustc --version
cargo install rustfmt
cargo install racer
rustup component add rust-src
cargo build
cargo run
```
