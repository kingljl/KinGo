# [YetAnotherFrameworkRust]

[YetAnotherFrameworkRust] is Another Framewor for web framework

##Hello world

```rust,no_run
#[macro_use] extern crate kRust;

use nickel::{kRust, HttpRouter};

fn main() {
    let mut server = kRust::new();
    server.get("**", middleware!("Hello World"));
    server.listen("127.0.0.1:6767");
}
```

### Dependencies

You'll need to create a *Cargo.toml* that looks like this;

```toml
[package]

name = "my-nickel-app"
version = "0.0.1"
authors = ["yourname"]

[dependencies.nickel]
version = "*"

