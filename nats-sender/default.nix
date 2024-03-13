{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule rec {
  pname = "nats-sender";
  version = "unstable";

  src = ./src;

  vendorHash = "sha256-oSVwA5QoKYIxFWxdqVnHtj2brsw9qYvx8ACY64+8GK4=";

  nativeBuildInputs = [ pkgs.musl pkgs.installShellFiles];

  CGO_ENABLED = 0;

  ldflags = [
    "-linkmode external"
    "-extldflags '-static -L${pkgs.musl}/lib'"
  ];

  postInstall = ''
    mv $out/bin/cmd $out/bin/nats-sender
    mkdir -p $out/config
    cp ./config/config.yml $out/config/

    # Completions will be added later
    # installShellCompletion --cmd nats-sender \
    #   --bash <($out/bin/nats-sender completion bash) \
    #   --zsh <($out/bin/nats-sender completion zsh) \
    #   --fish <($out/bin/nats-sender completion zsh)
  '';
}